package utils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveDocumentPath(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	root := filepath.Join(cwd, DocumentsDir)

	t.Run("valid nested path", func(t *testing.T) {
		got, err := ResolveDocumentPath("specifica/attivita/1/uuid-test.png")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(root, "specifica", "attivita", "1", "uuid-test.png")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("valid single segment", func(t *testing.T) {
		got, err := ResolveDocumentPath("file.pdf")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := filepath.Join(root, "file.pdf")
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		_, err := ResolveDocumentPath("")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "empty") {
			t.Fatalf("expected empty path error, got: %v", err)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		_, err := ResolveDocumentPath("   ")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects parent ..", func(t *testing.T) {
		_, err := ResolveDocumentPath("..")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects leading ..", func(t *testing.T) {
		_, err := ResolveDocumentPath("../etc/passwd")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects traversal in middle", func(t *testing.T) {
		_, err := ResolveDocumentPath("a/../../b")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects deep traversal (still has .. after clean)", func(t *testing.T) {
		// Three segments up from "1" land outside ordini_lavoro → path still contains ..
		_, err := ResolveDocumentPath("ordini_lavoro/allegati/1/../../../../esterno")
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("rejects only dot-dot segments", func(t *testing.T) {
		_, err := ResolveDocumentPath("../../../../z")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
