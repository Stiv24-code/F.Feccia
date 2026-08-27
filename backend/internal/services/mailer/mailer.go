// Package mailer ports OrderMesh's SMTP mailer: plain-text UTF-8 mails via
// net/smtp, with STARTTLS on submission ports and implicit TLS on 465, and
// the LOGIN-auth fallback required by smtp.office365.com. It implements the
// inboundorders.AcceptanceMailer seam — the confirmation mail sent when an
// operator accepts an inbound order.
package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"fratelli-feccia/config"
	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services/inboundorders"
)

// smtpTimeout bounds the whole SMTP exchange (dial + STARTTLS/TLS handshake
// + auth + DATA) — smtp.SendMail and tls.Dial have no timeout of their own,
// so a stalled server would otherwise hang the send forever.
const smtpTimeout = 30 * time.Second

type MailerService struct {
	cfg config.InboundConfig
}

// Compile-time check that the service satisfies the accept-flow seam.
var _ inboundorders.AcceptanceMailer = (*MailerService)(nil)

func NewMailerService(cfg config.InboundConfig) *MailerService {
	return &MailerService{cfg: cfg}
}

func (s *MailerService) Configured() bool { return s.cfg.SMTPConfigured() }

// SendAcceptance builds and sends the confirmation mail for an accepted
// inbound order, returning the actual recipient (TEST_RECIPIENT while
// ACCEPT_MODE=test, the order's sender in production).
func (s *MailerService) SendAcceptance(ctx context.Context, o dto.InboundOrderResponse) (string, error) {
	to, subject, body := acceptanceMail(s.cfg, o)
	if to == "" {
		if s.cfg.AcceptMode == "production" {
			return "", fmt.Errorf("l'ordine non ha un mittente (sender_email vuoto)")
		}
		return "", fmt.Errorf("TEST_RECIPIENT non configurato (ACCEPT_MODE=%s)", s.cfg.AcceptMode)
	}
	if err := s.Send(ctx, to, subject, body); err != nil {
		return "", err
	}
	return to, nil
}

// Send delivers a plain-text UTF-8 mail through the configured SMTP server.
// (net/smtp has no context support — ctx is accepted for interface symmetry.)
func (s *MailerService) Send(ctx context.Context, to, subject, body string) error {
	return s.send(ctx, to, subject, "text/plain", strings.ReplaceAll(body, "\n", "\r\n"))
}

// SendHTML delivers an HTML mail — used where the body needs a real
// clickable link (an <a href> — a bare URL in a text/plain body isn't
// always auto-linkified by the recipient's mail client) rather than
// plain multi-line text like the inbound-order acceptance mail above.
func (s *MailerService) SendHTML(ctx context.Context, to, subject, htmlBody string) error {
	return s.send(ctx, to, subject, "text/html", htmlBody)
}

// send delivers a UTF-8 mail through the configured SMTP server. Port 587
// (and any other) upgrades with STARTTLS when the server advertises it; port
// 465 needs an implicit-TLS dial. Both paths share a single dial+session
// deadline (smtpTimeout) — net/smtp has no context support, so this is the
// substitute for ctx cancellation. Traced as its own span — SMTP is a
// genuinely slow, blocking network call (this session's Brevo debugging
// alone: IP authorization, sender verification, DKIM/DMARC — all show up as
// "did the send hang or fail fast" in a trace).
func (s *MailerService) send(ctx context.Context, to, subject, contentType, body string) (err error) {
	_, span := otel.Tracer("fratelli-feccia/mailer").Start(ctx, "mailer.send",
		trace.WithAttributes(attribute.String("mail.content_type", contentType)),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	cfg := s.cfg
	if !cfg.SMTPConfigured() {
		return fmt.Errorf("SMTP non configurato (SMTP_HOST vuoto)")
	}
	from := cfg.MailFrom
	if from == "" {
		from = cfg.SMTPUser
	}
	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + mime.QEncoding.Encode("utf-8", subject),
		"MIME-Version: 1.0",
		"Content-Type: " + contentType + "; charset=utf-8",
		"",
		body,
	}, "\r\n")

	addr := cfg.SMTPHost + ":" + cfg.SMTPPort
	var auth smtp.Auth
	if cfg.SMTPUser != "" {
		auth = &autoAuth{user: cfg.SMTPUser, pass: cfg.SMTPPass, host: cfg.SMTPHost}
	}

	conn, err := net.DialTimeout("tcp", addr, smtpTimeout)
	if err != nil {
		return err
	}
	if err := conn.SetDeadline(time.Now().Add(smtpTimeout)); err != nil {
		conn.Close()
		return err
	}
	if cfg.SMTPPort == "465" {
		conn = tls.Client(conn, &tls.Config{ServerName: cfg.SMTPHost})
	}

	c, err := smtp.NewClient(conn, cfg.SMTPHost)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()
	if cfg.SMTPPort != "465" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: cfg.SMTPHost}); err != nil {
				return err
			}
		}
	}
	if auth != nil {
		if err := c.Auth(auth); err != nil {
			return err
		}
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}

// autoAuth picks the SASL mechanism the server actually advertises:
// PLAIN where available, otherwise LOGIN (required by e.g. smtp.office365.com,
// which answers "504 Unrecognized authentication type" to AUTH PLAIN).
type autoAuth struct {
	user, pass, host string
	inner            smtp.Auth
}

func (a *autoAuth) Start(s *smtp.ServerInfo) (string, []byte, error) {
	for _, m := range s.Auth {
		if strings.EqualFold(m, "PLAIN") {
			a.inner = smtp.PlainAuth("", a.user, a.pass, a.host)
			return a.inner.Start(s)
		}
	}
	if !s.TLS {
		return "", nil, fmt.Errorf("connessione non TLS: LOGIN auth rifiutata")
	}
	a.inner = nil
	return "LOGIN", nil, nil
}

func (a *autoAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if a.inner != nil {
		return a.inner.Next(fromServer, more)
	}
	if !more {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(string(fromServer))) {
	case "username:":
		return []byte(a.user), nil
	case "password:":
		return []byte(a.pass), nil
	default: // some servers use non-standard prompts; username comes first
		return []byte(a.user), nil
	}
}

// acceptanceMail builds the confirmation sent when an order is accepted.
// While ACCEPT_MODE != "production" every mail goes to TEST_RECIPIENT with
// the real recipient noted in the body.
func acceptanceMail(cfg config.InboundConfig, o dto.InboundOrderResponse) (to, subject, body string) {
	subject = fmt.Sprintf("Conferma ordine %s — %s", o.Ref, o.Client)
	kg := "—"
	if o.Kg > 0 {
		kg = fmt.Sprintf("%d kg", o.Kg)
	}
	lines := []string{
		"Buongiorno,",
		"",
		"confermiamo l'accettazione del seguente ordine di trasporto:",
		"",
		"Cliente:      " + o.Client,
		"Riferimento:  " + o.Ref,
		"Prodotto:     " + o.Product,
		"Quantità:     " + kg,
		"Carico:       " + o.LoadDate + " — " + o.LoadPlace,
		"Consegna:     " + o.DeliveryDate + " — " + o.DeliveryPlace,
	}
	if o.Rate != "" && o.Rate != "—" {
		lines = append(lines, "Nolo:         "+o.Rate)
	}
	lines = append(lines, "", "Cordiali saluti,", "Feccia F.lli S.r.l. — Order Desk")

	to = o.SenderEmail
	if cfg.AcceptMode != "production" {
		to = cfg.TestRecipient
		lines = append([]string{
			"[MODALITÀ TEST — destinatario reale: " + o.SenderEmail + "]",
			"",
		}, lines...)
	}
	return to, subject, strings.Join(lines, "\n")
}
