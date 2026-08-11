package mailer

import (
	"context"
	"net/smtp"
	"strings"
	"testing"

	"fratelli-feccia/config"
	"fratelli-feccia/internal/dto"
)

func baseOrder() dto.InboundOrderResponse {
	return dto.InboundOrderResponse{
		Client:        "ACME S.r.l.",
		SenderEmail:   "ordini@acme.it",
		Ref:           "ORD-42",
		Product:       "Melassa",
		Kg:            28000,
		LoadDate:      "2026-08-10",
		LoadPlace:     "Ravenna",
		DeliveryDate:  "2026-08-11",
		DeliveryPlace: "Verona",
		Rate:          "€ 950",
	}
}

func TestAcceptanceMail_TestModeRoutesToTestRecipient(t *testing.T) {
	cfg := config.InboundConfig{AcceptMode: "test", TestRecipient: "operatore@feccia.it"}
	to, subject, body := acceptanceMail(cfg, baseOrder())

	if to != "operatore@feccia.it" {
		t.Fatalf("expected test recipient, got %q", to)
	}
	if !strings.Contains(subject, "ORD-42") || !strings.Contains(subject, "ACME") {
		t.Fatalf("unexpected subject: %q", subject)
	}
	if !strings.Contains(body, "[MODALITÀ TEST — destinatario reale: ordini@acme.it]") {
		t.Fatalf("expected test-mode banner with real recipient, got:\n%s", body)
	}
	if !strings.Contains(body, "28000 kg") || !strings.Contains(body, "Nolo:         € 950") {
		t.Fatalf("expected kg and rate lines, got:\n%s", body)
	}
}

func TestAcceptanceMail_ProductionRoutesToSender(t *testing.T) {
	cfg := config.InboundConfig{AcceptMode: "production", TestRecipient: "operatore@feccia.it"}
	o := baseOrder()
	o.Kg = 0
	o.Rate = ""
	to, _, body := acceptanceMail(cfg, o)

	if to != "ordini@acme.it" {
		t.Fatalf("expected the order sender in production, got %q", to)
	}
	if strings.Contains(body, "MODALITÀ TEST") {
		t.Fatalf("expected no test banner in production, got:\n%s", body)
	}
	if !strings.Contains(body, "Quantità:     —") {
		t.Fatalf("expected em-dash for zero kg, got:\n%s", body)
	}
	if strings.Contains(body, "Nolo:") {
		t.Fatalf("expected no rate line when empty, got:\n%s", body)
	}
}

func TestSendAcceptance_EmptyRecipientFailsClearly(t *testing.T) {
	// Test mode with no TEST_RECIPIENT: fail with a clear message instead of
	// handing an empty RCPT TO to the SMTP server.
	svc := NewMailerService(config.InboundConfig{
		SMTPHost:   "smtp.example.com",
		AcceptMode: "test",
	})
	_, err := svc.SendAcceptance(context.Background(), baseOrder())
	if err == nil || !strings.Contains(err.Error(), "TEST_RECIPIENT") {
		t.Fatalf("expected TEST_RECIPIENT error, got %v", err)
	}

	// Production with an order that has no sender address.
	svc = NewMailerService(config.InboundConfig{
		SMTPHost:   "smtp.example.com",
		AcceptMode: "production",
	})
	o := baseOrder()
	o.SenderEmail = ""
	_, err = svc.SendAcceptance(context.Background(), o)
	if err == nil || !strings.Contains(err.Error(), "mittente") {
		t.Fatalf("expected missing-sender error, got %v", err)
	}
}

func TestSend_NotConfigured(t *testing.T) {
	svc := NewMailerService(config.InboundConfig{})
	if svc.Configured() {
		t.Fatalf("expected Configured()==false with no SMTP_HOST")
	}
	if err := svc.Send(context.Background(), "a@b.it", "s", "b"); err == nil {
		t.Fatalf("expected error when SMTP is not configured")
	}
}

func TestAutoAuth_PicksPlainWhenAdvertised(t *testing.T) {
	a := &autoAuth{user: "u@example.com", pass: "p", host: "smtp.example.com"}
	mech, _, err := a.Start(&smtp.ServerInfo{
		Name: "smtp.example.com",
		TLS:  true,
		Auth: []string{"LOGIN", "PLAIN"},
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if mech != "PLAIN" {
		t.Fatalf("expected PLAIN when advertised, got %q", mech)
	}
}

func TestAutoAuth_FallsBackToLoginOverTLS(t *testing.T) {
	// office365 advertises LOGIN only.
	a := &autoAuth{user: "u@example.com", pass: "p", host: "smtp.office365.com"}
	mech, _, err := a.Start(&smtp.ServerInfo{
		Name: "smtp.office365.com",
		TLS:  true,
		Auth: []string{"LOGIN", "XOAUTH2"},
	})
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if mech != "LOGIN" {
		t.Fatalf("expected LOGIN fallback, got %q", mech)
	}

	// The LOGIN dialogue: username prompt, then password prompt.
	if got, _ := a.Next([]byte("Username:"), true); string(got) != "u@example.com" {
		t.Fatalf("expected username answer, got %q", got)
	}
	if got, _ := a.Next([]byte("Password:"), true); string(got) != "p" {
		t.Fatalf("expected password answer, got %q", got)
	}
}

func TestAutoAuth_RefusesLoginWithoutTLS(t *testing.T) {
	a := &autoAuth{user: "u", pass: "p", host: "h"}
	if _, _, err := a.Start(&smtp.ServerInfo{Name: "h", TLS: false, Auth: []string{"LOGIN"}}); err == nil {
		t.Fatalf("expected refusal of LOGIN auth on a non-TLS connection")
	}
}
