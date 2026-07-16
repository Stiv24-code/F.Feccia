// Package pdfgen ports backend/cmr_pdf.py, backend/instructions_pdf.py and
// backend/invoice_pdf.py using go-pdf/fpdf — the Go continuation of the same
// PHP-derived FPDF lineage as Python's fpdf2, so the cell/multi-cell,
// absolute-coordinate drawing model translates almost 1:1.
package pdfgen

import (
	"strconv"
	"strings"
	"time"
)

// safe mirrors the Python generators' `_safe()`: go-pdf/fpdf's core fonts
// (Helvetica/Times/Courier) only render Latin-1 code points when fed raw
// bytes in that range — there is no UTF-8 font loaded here (that would
// require shipping .ttf files in the Docker image). Runes outside Latin-1
// become '?', matching Python's `errors="replace"`.
func safe(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	b := make([]byte, len(runes))
	for i, r := range runes {
		if r > 255 {
			b[i] = '?'
		} else {
			b[i] = byte(r)
		}
	}
	return string(b)
}

// fmtDate mirrors `_fmt_date`: ISO "YYYY-MM-DD..." -> "DD/MM/YYYY".
func fmtDate(value string) string {
	if value == "" {
		return ""
	}
	v := value
	if len(v) > 10 {
		v = v[:10]
	}
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		return value
	}
	return t.Format("02/01/2006")
}

// fmtTimeWindow mirrors `_fmt_time_window`.
func fmtTimeWindow(da, a string) string {
	da = strings.TrimSpace(da)
	a = strings.TrimSpace(a)
	if da != "" && a != "" {
		return da + " - " + a
	}
	if da != "" {
		return da
	}
	return a
}

// fmtEuro mirrors invoice_pdf.py's `_euro`: Italian-locale grouping (dot for
// thousands, comma for decimals) with a trailing "EUR" literal, not the €
// glyph (Latin-1 core fonts can't render it without a Windows-1252 code page
// or a loaded Unicode font).
func fmtEuro(value float64) string {
	neg := value < 0
	if neg {
		value = -value
	}
	cents := int64(value*100 + 0.5)
	intPart := cents / 100
	decPart := cents % 100

	digits := strconv.FormatInt(intPart, 10)
	var grouped strings.Builder
	for i, d := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 {
			grouped.WriteByte('.')
		}
		grouped.WriteRune(d)
	}

	sign := ""
	if neg {
		sign = "-"
	}
	return sign + grouped.String() + "," + fmt02(decPart) + " EUR"
}

func fmt02(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) < 2 {
		return "0" + s
	}
	return s
}

// fmtG mirrors Python's `f"{value:g}"`: minimal decimal representation,
// dropping trailing zeros/the decimal point entirely for whole numbers.
func fmtG(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}
