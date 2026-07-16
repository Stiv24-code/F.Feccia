package s3invoices

import (
	"context"
	"testing"

	"fratelli-feccia/config"
)

func TestNewClient_DisabledWhenBucketEmpty(t *testing.T) {
	c, err := NewClient(context.Background(), config.S3Config{InvoicesBucket: ""})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if c.IsEnabled() {
		t.Fatal("expected IsEnabled() == false when bucket is empty")
	}
}

func TestUploadInvoicePDF_NoopWhenDisabled(t *testing.T) {
	c, _ := NewClient(context.Background(), config.S3Config{})
	result, err := c.UploadInvoicePDF(context.Background(), []byte("%PDF-fake"), "invoices/2026/x.pdf", nil)
	if err != nil {
		t.Fatalf("expected no error when disabled, got %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil result when disabled, got %+v", result)
	}
}

func TestGetPresignedURL_ErrorsWhenDisabled(t *testing.T) {
	c, _ := NewClient(context.Background(), config.S3Config{})
	_, err := c.GetPresignedURL(context.Background(), "invoices/2026/x.pdf", 0)
	if err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestFetchInvoicePDF_ErrorsWhenDisabled(t *testing.T) {
	c, _ := NewClient(context.Background(), config.S3Config{})
	_, err := c.FetchInvoicePDF(context.Background(), "invoices/2026/x.pdf")
	if err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestBuildInvoiceKey(t *testing.T) {
	cases := []struct {
		invoiceID, numero, year, want string
	}{
		{"abc-123", "O/F-26/0001", "2026", "invoices/2026/O-F-26-0001-abc-123.pdf"},
		{"abc-123", "", "2026", "invoices/2026/draft-abc-123.pdf"},
	}
	for _, tc := range cases {
		got := BuildInvoiceKey(tc.invoiceID, tc.numero, tc.year)
		if got != tc.want {
			t.Errorf("BuildInvoiceKey(%q, %q, %q) = %q, want %q", tc.invoiceID, tc.numero, tc.year, got, tc.want)
		}
	}
}

func TestAsciiSanitize(t *testing.T) {
	cases := []struct{ in, want string }{
		{"plain ascii", "plain ascii"},
		{"café €5", "caf? ?5"},
	}
	for _, tc := range cases {
		if got := asciiSanitize(tc.in); got != tc.want {
			t.Errorf("asciiSanitize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSanitizeMetadata_TruncatesAndSanitizes(t *testing.T) {
	long := make([]byte, 2000)
	for i := range long {
		long[i] = 'a'
	}
	meta := sanitizeMetadata(map[string]string{"note": "café", "long": string(long)})
	if meta["note"] != "caf?" {
		t.Fatalf("expected sanitized note, got %q", meta["note"])
	}
	if len(meta["long"]) != 1024 {
		t.Fatalf("expected truncation to 1024 chars, got %d", len(meta["long"]))
	}
}

func TestSanitizeMetadata_NilForEmptyInput(t *testing.T) {
	if got := sanitizeMetadata(nil); got != nil {
		t.Fatalf("expected nil for empty metadata, got %+v", got)
	}
}
