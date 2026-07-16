package pdfgen

import "testing"

func TestFmtEuro(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0,00 EUR"},
		{1234.5, "1.234,50 EUR"},
		{1234567.89, "1.234.567,89 EUR"},
		{9.999, "10,00 EUR"},
		{-42.5, "-42,50 EUR"},
		{100, "100,00 EUR"},
	}
	for _, tc := range cases {
		if got := fmtEuro(tc.in); got != tc.want {
			t.Errorf("fmtEuro(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFmtDate(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"2026-03-15", "15/03/2026"},
		{"2026-03-15T10:30:00Z", "15/03/2026"},
		{"not-a-date", "not-a-date"},
	}
	for _, tc := range cases {
		if got := fmtDate(tc.in); got != tc.want {
			t.Errorf("fmtDate(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFmtTimeWindow(t *testing.T) {
	cases := []struct {
		da, a, want string
	}{
		{"08:00", "12:00", "08:00 - 12:00"},
		{"08:00", "", "08:00"},
		{"", "12:00", "12:00"},
		{"", "", ""},
	}
	for _, tc := range cases {
		if got := fmtTimeWindow(tc.da, tc.a); got != tc.want {
			t.Errorf("fmtTimeWindow(%q, %q) = %q, want %q", tc.da, tc.a, got, tc.want)
		}
	}
}

func TestSafe_ReplacesNonLatin1Runes(t *testing.T) {
	got := safe("café €5 日本語")
	// 'é' (U+00E9) is within Latin-1 and must pass through; '€' and the
	// Japanese characters are outside Latin-1 and must become '?'.
	want := "caf\xe9 ?5 ???"
	if got != want {
		t.Errorf("safe() = %q (%v), want %q (%v)", got, []byte(got), want, []byte(want))
	}
}

func TestFmtG(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{5, "5"},
		{2.5, "2.5"},
		{2.50, "2.5"},
	}
	for _, tc := range cases {
		if got := fmtG(tc.in); got != tc.want {
			t.Errorf("fmtG(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
