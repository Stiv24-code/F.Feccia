// Package s3invoices ports backend/s3_invoices.py: invoice PDF archival to
// S3 with Object Lock (COMPLIANCE mode, 10-year retention matching the
// Italian fiscal record-keeping obligation) + presigned download URLs.
//
// No-op when the bucket isn't configured (IsEnabled() == false) — the same
// dev-mode behavior as the Python original, since the bucket isn't
// provisioned yet (see infra/aws/provision.sh).
package s3invoices

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	"fratelli-feccia/config"
)

var ErrNotConfigured = errors.New("S3 invoices bucket non configurato")

type Client struct {
	s3             *s3.Client
	presign        *s3.PresignClient
	bucket         string
	region         string
	retentionYears int
	presignedTTL   time.Duration
}

// NewClient mirrors s3_invoices.py's lazy `_client()`: credentials are
// resolved by the SDK's default provider chain (IAM instance profile on
// EC2, or AWS_ACCESS_KEY_ID/AWS_SECRET_ACCESS_KEY env vars in dev) — no
// explicit handling here. If the bucket isn't configured, no AWS client is
// even constructed.
func NewClient(ctx context.Context, cfg config.S3Config) (*Client, error) {
	c := &Client{
		bucket:         cfg.InvoicesBucket,
		region:         cfg.Region,
		retentionYears: cfg.InvoicesRetentionYears,
		presignedTTL:   time.Duration(cfg.PresignedTTLSeconds) * time.Second,
	}
	if !c.IsEnabled() {
		return c, nil
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}
	c.s3 = s3.NewFromConfig(awsCfg)
	c.presign = s3.NewPresignClient(c.s3)
	return c, nil
}

// IsEnabled mirrors is_enabled: true only if the bucket is configured.
func (c *Client) IsEnabled() bool {
	return c.bucket != ""
}

// BuildInvoiceKey mirrors build_invoice_key: `invoices/{year}/{numero-slug}-{invoice_id}.pdf`.
func BuildInvoiceKey(invoiceID, numero, year string) string {
	if year == "" {
		year = time.Now().UTC().Format("2006")
	}
	if len(year) > 4 {
		year = year[:4]
	}
	safeNumero := numero
	if safeNumero == "" {
		safeNumero = "draft"
	}
	safeNumero = strings.ReplaceAll(safeNumero, "/", "-")
	safeNumero = strings.ReplaceAll(safeNumero, " ", "_")
	return fmt.Sprintf("invoices/%s/%s-%s.pdf", year, safeNumero, invoiceID)
}

type UploadResult struct {
	Key         string
	UploadedAt  string
	RetainUntil string
}

// UploadInvoicePDF mirrors upload_invoice_pdf: PUT with Object Lock
// COMPLIANCE mode at the configured retention. Returns nil (not an error)
// when disabled — the caller must treat that as "PDF not archived", exactly
// like Python's empty-dict return.
func (c *Client) UploadInvoicePDF(ctx context.Context, pdfBytes []byte, key string, metadata map[string]string) (*UploadResult, error) {
	if !c.IsEnabled() {
		return nil, nil
	}

	retainUntil := time.Now().UTC().AddDate(c.retentionYears, 0, 0)
	safeMetadata := sanitizeMetadata(metadata)

	_, err := c.s3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:                    &c.bucket,
		Key:                       &key,
		Body:                      bytes.NewReader(pdfBytes),
		ContentType:               strPtr("application/pdf"),
		ServerSideEncryption:      types.ServerSideEncryptionAes256,
		ObjectLockMode:            types.ObjectLockModeCompliance,
		ObjectLockRetainUntilDate: &retainUntil,
		Metadata:                  safeMetadata,
	})
	if err != nil {
		return nil, err
	}

	uploadedAt := time.Now().UTC()
	return &UploadResult{
		Key:         key,
		UploadedAt:  uploadedAt.Format(time.RFC3339),
		RetainUntil: retainUntil.Format(time.RFC3339),
	}, nil
}

// GetPresignedURL mirrors get_invoice_presigned_url: short-lived GET URL
// (default from config, override via ttl if > 0).
func (c *Client) GetPresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if !c.IsEnabled() {
		return "", ErrNotConfigured
	}
	if ttl <= 0 {
		ttl = c.presignedTTL
	}
	req, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &c.bucket, Key: &key,
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

// FetchInvoicePDF mirrors fetch_invoice_pdf: downloads the archived PDF.
func (c *Client) FetchInvoicePDF(ctx context.Context, key string) ([]byte, error) {
	if !c.IsEnabled() {
		return nil, ErrNotConfigured
	}
	out, err := c.s3.GetObject(ctx, &s3.GetObjectInput{Bucket: &c.bucket, Key: &key})
	if err != nil {
		return nil, err
	}
	defer out.Body.Close()
	return io.ReadAll(out.Body)
}

// sanitizeMetadata mirrors upload_invoice_pdf's metadata sanitization: S3
// object metadata only accepts ASCII printable characters.
func sanitizeMetadata(metadata map[string]string) map[string]string {
	if len(metadata) == 0 {
		return nil
	}
	out := make(map[string]string, len(metadata))
	for k, v := range metadata {
		out[k] = asciiSanitize(v)
	}
	return out
}

func asciiSanitize(s string) string {
	if len(s) > 1024 {
		s = s[:1024]
	}
	runes := []rune(s)
	b := make([]byte, len(runes))
	for i, r := range runes {
		if r > 127 {
			b[i] = '?'
		} else {
			b[i] = byte(r)
		}
	}
	return string(b)
}

func strPtr(s string) *string { return &s }
