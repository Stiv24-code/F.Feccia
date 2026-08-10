package mailscraper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Microsoft Graph backend for Exchange Online mailboxes, where Microsoft has
// permanently disabled password (basic) authentication for IMAP. Two modes:
//   - delegated: the user signs in once (browser or device code, via the
//     graphlogin CLI) and the service stores/refreshes the tokens in
//     DATA_DIR/graph_token.json;
//   - app-only: with GRAPH_CLIENT_SECRET set the service authenticates as
//     the application itself (client credentials) — no sign-in, no token
//     file, Conditional Access does not apply.

// defaultGraphClientID is Microsoft's public Azure CLI app registration,
// usable for delegated Graph scopes without registering a custom app.
// Override with GRAPH_CLIENT_ID if the tenant restricts it.
const defaultGraphClientID = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"

const graphScope = "https://graph.microsoft.com/Mail.ReadWrite offline_access"

type graphTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (s *MailScraperService) clientID() string {
	if s.cfg.GraphClientID != "" {
		return s.cfg.GraphClientID
	}
	return defaultGraphClientID
}

func (s *MailScraperService) tokenPath() string {
	return filepath.Join(s.cfg.DataDir, "graph_token.json")
}

func (s *MailScraperService) loggedIn() bool {
	_, err := os.Stat(s.tokenPath())
	return err == nil
}

func (s *MailScraperService) appOnly() bool { return s.cfg.GraphClientSecret != "" }

func (s *MailScraperService) loginURL(endpoint string) string {
	return "https://login.microsoftonline.com/" + s.cfg.GraphTenant + "/oauth2/v2.0/" + endpoint
}

// ── Delegated sign-in (used by cmd/graphlogin) ─────────────────────────────

// LoginBrowser runs the OAuth2 authorization-code flow with PKCE: it opens
// the system browser on the Microsoft sign-in page and receives the code on
// a loopback listener. Conditional Access policies that block the
// device-code flow normally allow this one — it is a regular browser
// sign-in. Requires the redirect URI http://localhost:8400/callback on the
// app registration (platform "Mobile and desktop applications").
const loopbackAddr = "127.0.0.1:8400"
const redirectURI = "http://localhost:8400/callback"

func (s *MailScraperService) LoginBrowser() error {
	verifier := randomURLSafe(48)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])
	state := randomURLSafe(16)

	ln, err := net.Listen("tcp", loopbackAddr)
	if err != nil {
		return fmt.Errorf("porta 8400 occupata: %w", err)
	}

	type result struct {
		code string
		err  error
	}
	resCh := make(chan result, 1)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if q.Get("state") != state || q.Get("code") == "" {
			fmt.Fprint(w, "<h3>Accesso non riuscito</h3><p>"+q.Get("error_description")+"</p>")
			resCh <- result{err: fmt.Errorf("%s — %s", q.Get("error"), q.Get("error_description"))}
			return
		}
		fmt.Fprint(w, "<h3>Accesso completato ✓</h3><p>Puoi chiudere questa finestra.</p>")
		resCh <- result{code: q.Get("code")}
	})}
	go srv.Serve(ln)
	defer srv.Shutdown(context.Background())

	authURL := s.loginURL("authorize") + "?" + url.Values{
		"client_id":             {s.clientID()},
		"response_type":         {"code"},
		"redirect_uri":          {redirectURI},
		"response_mode":         {"query"},
		"scope":                 {graphScope},
		"state":                 {state},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}.Encode()

	fmt.Println()
	fmt.Println("Apro il browser per l'accesso Microsoft…")
	fmt.Println("Se non si apre, copia questo link:")
	fmt.Println(authURL)
	fmt.Println()
	openBrowser(authURL)

	select {
	case res := <-resCh:
		if res.err != nil {
			return res.err
		}
		return s.exchangeCode(res.code, verifier)
	case <-time.After(10 * time.Minute):
		return fmt.Errorf("login: nessuna risposta entro 10 minuti, riprova")
	}
}

func openBrowser(u string) {
	switch runtime.GOOS {
	case "windows":
		_ = exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
	case "darwin":
		_ = exec.Command("open", u).Start()
	default:
		_ = exec.Command("xdg-open", u).Start()
	}
}

func (s *MailScraperService) exchangeCode(code, verifier string) error {
	resp, err := http.PostForm(s.loginURL("token"), url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {s.clientID()},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	})
	if err != nil {
		return err
	}
	var tk struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := decodeBody(resp, &tk); err != nil {
		return err
	}
	if tk.Error != "" {
		return fmt.Errorf("token: %s — %s", tk.Error, tk.ErrorDesc)
	}
	return s.saveTokens(graphTokens{
		AccessToken:  tk.AccessToken,
		RefreshToken: tk.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tk.ExpiresIn) * time.Second),
	})
}

func randomURLSafe(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// LoginDevice runs the device-code flow and stores the tokens — for hosts
// where opening a browser is impossible and Conditional Access allows it.
func (s *MailScraperService) LoginDevice() error {
	resp, err := http.PostForm(s.loginURL("devicecode"), url.Values{
		"client_id": {s.clientID()},
		"scope":     {graphScope},
	})
	if err != nil {
		return err
	}
	var dc struct {
		DeviceCode      string `json:"device_code"`
		UserCode        string `json:"user_code"`
		VerificationURI string `json:"verification_uri"`
		ExpiresIn       int    `json:"expires_in"`
		Interval        int    `json:"interval"`
		Message         string `json:"message"`
		Error           string `json:"error"`
		ErrorDesc       string `json:"error_description"`
	}
	if err := decodeBody(resp, &dc); err != nil {
		return err
	}
	if dc.Error != "" {
		return fmt.Errorf("devicecode: %s — %s", dc.Error, dc.ErrorDesc)
	}

	fmt.Println()
	fmt.Println("  ┌─────────────────────────────────────────────────────┐")
	fmt.Printf("  │  1. Apri   %s\n", dc.VerificationURI)
	fmt.Printf("  │  2. Codice %s\n", dc.UserCode)
	fmt.Println("  │  3. Accedi con l'account della casella ordini        ")
	fmt.Println("  └─────────────────────────────────────────────────────┘")
	fmt.Println()

	interval := time.Duration(max(dc.Interval, 5)) * time.Second
	deadline := time.Now().Add(time.Duration(dc.ExpiresIn) * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(interval)
		resp, err := http.PostForm(s.loginURL("token"), url.Values{
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
			"client_id":   {s.clientID()},
			"device_code": {dc.DeviceCode},
		})
		if err != nil {
			return err
		}
		var tk struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
			ExpiresIn    int    `json:"expires_in"`
			Error        string `json:"error"`
			ErrorDesc    string `json:"error_description"`
		}
		if err := decodeBody(resp, &tk); err != nil {
			return err
		}
		switch tk.Error {
		case "":
			return s.saveTokens(graphTokens{
				AccessToken:  tk.AccessToken,
				RefreshToken: tk.RefreshToken,
				ExpiresAt:    time.Now().Add(time.Duration(tk.ExpiresIn) * time.Second),
			})
		case "authorization_pending":
			continue
		case "slow_down":
			interval += 5 * time.Second
		default:
			return fmt.Errorf("login: %s — %s", tk.Error, tk.ErrorDesc)
		}
	}
	return fmt.Errorf("login: codice scaduto, riprova")
}

func (s *MailScraperService) saveTokens(t graphTokens) error {
	if err := os.MkdirAll(s.cfg.DataDir, 0o755); err != nil {
		return err
	}
	raw, _ := json.MarshalIndent(t, "", " ")
	if err := os.WriteFile(s.tokenPath(), raw, 0o600); err != nil {
		return err
	}
	slog.Info("login Graph completato", "token_path", s.tokenPath())
	return nil
}

// ── Tokens at scrape time ──────────────────────────────────────────────────

// graphAppToken authenticates as the application itself (client credentials).
// No user sign-in involved, so Conditional Access policies do not apply.
func (s *MailScraperService) graphAppToken() (string, error) {
	s.appTok.Lock()
	defer s.appTok.Unlock()
	if s.appTok.val != "" && time.Until(s.appTok.exp) > 2*time.Minute {
		return s.appTok.val, nil
	}
	resp, err := http.PostForm(s.loginURL("token"), url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {s.clientID()},
		"client_secret": {s.cfg.GraphClientSecret},
		"scope":         {"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return "", err
	}
	var tk struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Error       string `json:"error"`
		ErrorDesc   string `json:"error_description"`
	}
	if err := decodeBody(resp, &tk); err != nil {
		return "", err
	}
	if tk.Error != "" {
		return "", fmt.Errorf("client credentials: %s — %s", tk.Error, tk.ErrorDesc)
	}
	s.appTok.val = tk.AccessToken
	s.appTok.exp = time.Now().Add(time.Duration(tk.ExpiresIn) * time.Second)
	return s.appTok.val, nil
}

// graphToken picks the right credential: app-only when a client secret is
// configured, otherwise the delegated tokens saved by cmd/graphlogin.
func (s *MailScraperService) graphToken() (string, error) {
	if s.appOnly() {
		return s.graphAppToken()
	}
	return s.graphAccessToken()
}

// graphMailboxBase is the Graph resource root: /users/{mailbox} in app-only
// mode (app tokens cannot use /me), /me with delegated tokens.
func (s *MailScraperService) graphMailboxBase() (string, error) {
	if s.appOnly() {
		if s.cfg.GraphMailbox == "" {
			return "", fmt.Errorf("GRAPH_MAILBOX mancante: indica la casella da leggere in .env")
		}
		return "https://graph.microsoft.com/v1.0/users/" + url.PathEscape(s.cfg.GraphMailbox), nil
	}
	return "https://graph.microsoft.com/v1.0/me", nil
}

// graphAccessToken returns a valid delegated token, refreshing near expiry.
func (s *MailScraperService) graphAccessToken() (string, error) {
	raw, err := os.ReadFile(s.tokenPath())
	if err != nil {
		return "", fmt.Errorf("nessun token Graph: esegui prima `go run ./cmd/graphlogin`")
	}
	var t graphTokens
	if err := json.Unmarshal(raw, &t); err != nil {
		return "", err
	}
	if time.Until(t.ExpiresAt) > 2*time.Minute {
		return t.AccessToken, nil
	}
	resp, err := http.PostForm(s.loginURL("token"), url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {s.clientID()},
		"refresh_token": {t.RefreshToken},
		"scope":         {graphScope},
	})
	if err != nil {
		return "", err
	}
	var tk struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	if err := decodeBody(resp, &tk); err != nil {
		return "", err
	}
	if tk.Error != "" {
		return "", fmt.Errorf("refresh token: %s — %s (riesegui `go run ./cmd/graphlogin`)", tk.Error, tk.ErrorDesc)
	}
	if tk.RefreshToken == "" {
		tk.RefreshToken = t.RefreshToken
	}
	if err := s.saveTokens(graphTokens{
		AccessToken:  tk.AccessToken,
		RefreshToken: tk.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Duration(tk.ExpiresIn) * time.Second),
	}); err != nil {
		return "", err
	}
	return tk.AccessToken, nil
}

// ── Scraping via Graph ─────────────────────────────────────────────────────

// scrapeGraph reads unread inbox messages, parses those whose subject
// contains the subject filter and marks the examined ones as read.
func (s *MailScraperService) scrapeGraph(ctx context.Context) (added, scanned int, err error) {
	token, err := s.graphToken()
	if err != nil {
		return 0, 0, err
	}
	base, err := s.graphMailboxBase()
	if err != nil {
		return 0, 0, err
	}

	q := base + "/mailFolders/inbox/messages" +
		"?$filter=isRead%20eq%20false&$top=50&$select=id,subject,from,body"
	req, err := http.NewRequestWithContext(ctx, "GET", q, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Prefer", `outlook.body-content-type="text"`)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0, err
	}
	var list struct {
		Value []struct {
			ID      string `json:"id"`
			Subject string `json:"subject"`
			From    struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			} `json:"from"`
			Body struct {
				Content string `json:"content"`
			} `json:"body"`
		} `json:"value"`
		Error *struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := decodeBody(resp, &list); err != nil {
		return 0, 0, err
	}
	if list.Error != nil {
		return 0, 0, fmt.Errorf("graph: %s — %s", list.Error.Code, list.Error.Message)
	}

	for _, m := range list.Value {
		if !strings.Contains(m.Subject, s.cfg.SubjectFilter) {
			continue // normal mail: leave it unread and untouched
		}
		scanned++
		order, ok := parseOrderBody(m.Body.Content, m.From.EmailAddress.Address)
		if !ok {
			slog.Info("scrape: mail ignorata (formato non riconosciuto)", "subject", m.Subject)
		} else if s.store(ctx, order) {
			added++
		}
		s.markRead(ctx, base, token, m.ID)
	}
	return added, scanned, nil
}

func (s *MailScraperService) markRead(ctx context.Context, base, token, id string) {
	body := strings.NewReader(`{"isRead": true}`)
	req, err := http.NewRequestWithContext(ctx, "PATCH", base+"/messages/"+url.PathEscape(id), body)
	if err != nil {
		return
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Warn("scrape: impossibile marcare come letta", "error", err)
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

func decodeBody(resp *http.Response, v any) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(v)
}
