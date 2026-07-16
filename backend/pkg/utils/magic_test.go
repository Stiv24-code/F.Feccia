package utils

import (
	"testing"
)

// minimal PDF body: magic check only requires %PDF prefix
var minimalPDF = []byte("%PDF-1.4\n%minimal test object\n%%EOF\n")

func TestValidateMagicBytes(t *testing.T) {
	t.Run("pdf ok", func(t *testing.T) {
		if err := ValidateMagicBytes(".pdf", minimalPDF); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("pdf rejects html", func(t *testing.T) {
		if err := ValidateMagicBytes(".pdf", []byte("<!DOCTYPE html><html>")); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("png ok", func(t *testing.T) {
		png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
		if err := ValidateMagicBytes(".png", png); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("jpeg ok", func(t *testing.T) {
		j := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}
		if err := ValidateMagicBytes(".jpg", j); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("gif ok", func(t *testing.T) {
		if err := ValidateMagicBytes(".gif", []byte("GIF89a\000\000")); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("docx zip ok", func(t *testing.T) {
		// Local file header
		b := []byte{'P', 'K', 0x03, 0x04, 0, 0, 0, 0}
		if err := ValidateMagicBytes(".docx", b); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("xlsx zip ok", func(t *testing.T) {
		b := []byte{'P', 'K', 0x03, 0x04, 0, 0, 0, 0}
		if err := ValidateMagicBytes(".xlsx", b); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("docx rejects without PK", func(t *testing.T) {
		if err := ValidateMagicBytes(".docx", []byte("<?xml ")); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ole doc ok", func(t *testing.T) {
		b := make([]byte, 16)
		copy(b, []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1})
		if err := ValidateMagicBytes(".doc", b); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("empty", func(t *testing.T) {
		if err := ValidateMagicBytes(".pdf", nil); err == nil {
			t.Fatal("expected error")
		}
	})
}
