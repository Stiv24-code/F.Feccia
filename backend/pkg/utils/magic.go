package utils

import (
	"bytes"
	"fmt"
	"strings"
)

// Min bytes needed to read from the start of a file for magic-number checks.
const magicHeaderSize = 512

// ErrMagicMismatch is returned when file body does not match the declared extension (HTTP 422).
const errMagicMismatch = "File content does not match the declared type"

var (
	pngHeader = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	oleHeader = []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1} // legacy .doc / .xls (CFB)
)

// ValidateMagicBytes checks that the beginning of the file matches the expected format for ext.
// ext must be lowercase and include the dot (e.g. ".pdf").
func ValidateMagicBytes(ext string, data []byte) error {
	ext = strings.ToLower(ext)
	if len(data) == 0 {
		return NewAPIError(422, errMagicMismatch)
	}

	switch ext {
	case ".pdf":
		if !bytes.HasPrefix(data, []byte("%PDF")) {
			return NewAPIError(422, errMagicMismatch)
		}
	case ".jpg", ".jpeg":
		if len(data) < 3 || data[0] != 0xFF || data[1] != 0xD8 || data[2] != 0xFF {
			return NewAPIError(422, errMagicMismatch)
		}
	case ".png":
		if len(data) < len(pngHeader) || !bytes.Equal(data[:len(pngHeader)], pngHeader) {
			return NewAPIError(422, errMagicMismatch)
		}
	case ".gif":
		if len(data) < 6 {
			return NewAPIError(422, errMagicMismatch)
		}
		if !bytes.HasPrefix(data, []byte("GIF87a")) && !bytes.HasPrefix(data, []byte("GIF89a")) {
			return NewAPIError(422, errMagicMismatch)
		}
	case ".doc", ".xls":
		if len(data) < len(oleHeader) || !bytes.Equal(data[:len(oleHeader)], oleHeader) {
			return NewAPIError(422, errMagicMismatch)
		}
	case ".docx", ".xlsx":
		// OOXML is a ZIP archive; reject HTML/scripts masquerading as Office.
		if len(data) < 4 || data[0] != 'P' || data[1] != 'K' {
			return NewAPIError(422, errMagicMismatch)
		}
		// Local file header signature (normal Office documents)
		if data[2] != 0x03 || data[3] != 0x04 {
			return NewAPIError(422, errMagicMismatch)
		}
	default:
		return fmt.Errorf("validateMagicBytes: unsupported ext %q", ext)
	}
	return nil
}
