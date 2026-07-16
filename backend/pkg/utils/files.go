package utils

import (
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const DocumentsDir = "documents"

func EnsureDocumentsDir() error {
	return os.MkdirAll(DocumentsDir, 0755)
}

// ResolveDocumentPath returns an absolute path under DocumentsDir for a relative key (e.g. "specifica/attivita/1/uuid.png").
// Rejects path traversal and paths outside the documents root.
func ResolveDocumentPath(rel string) (string, error) {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return "", fmt.Errorf("empty path")
	}
	clean := filepath.Clean(filepath.FromSlash(rel))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("invalid path")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	root := filepath.Join(cwd, DocumentsDir)
	full := filepath.Join(root, clean)
	relToRoot, err := filepath.Rel(root, full)
	if err != nil {
		return "", err
	}
	if relToRoot == ".." || strings.HasPrefix(relToRoot, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return full, nil
}

const MaxFileSize = 10 * 1024 * 1024 // Increased to 10 MB for larger documents if needed, but controlled

// AllowedExtensions defines which files can be uploaded
var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
}

func SaveFile(fileHeader *multipart.FileHeader, subDir string) (string, error) {
	if fileHeader.Size > MaxFileSize {
		return "", fmt.Errorf("file too large: max size is %d bytes", MaxFileSize)
	}

	// 1. Validate Extension (Unrestricted Upload Protection)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !AllowedExtensions[ext] {
		return "", NewAPIError(422, "Unsupported file type")
	}

	cwd, _ := os.Getwd()
	// 2. Sanitise subDir (Path Traversal Protection)
	cleanSubDir := filepath.Clean(subDir)
	if strings.Contains(cleanSubDir, "..") {
		return "", fmt.Errorf("invalid subdirectory path")
	}

	baseDir := filepath.Join(cwd, DocumentsDir, cleanSubDir)

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", fmt.Errorf("can't create dir %s: %w", baseDir, err)
	}

	// 3. Generate UUID Filename (Predictability & Path Traversal Protection)
	// We keep the original extension but replace the filename with a UUID
	uniqueName := uuid.New().String() + ext
	destPath := filepath.Join(baseDir, uniqueName)

	// save file (read header for magic-byte validation before persisting)
	srcFile, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("can't open uploaded file: %w", err)
	}
	defer srcFile.Close()

	head := make([]byte, magicHeaderSize)
	n, err := io.ReadFull(srcFile, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("can't read uploaded file: %w", err)
	}
	if err := ValidateMagicBytes(ext, head[:n]); err != nil {
		return "", err
	}

	dstFile, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("can't create file %s: %w", destPath, err)
	}
	defer func() { _ = dstFile.Close() }()

	if _, werr := dstFile.Write(head[:n]); werr != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("can't save file %s: %w", destPath, werr)
	}
	if _, cerr := io.Copy(dstFile, srcFile); cerr != nil {
		_ = os.Remove(destPath)
		return "", fmt.Errorf("can't save file %s: %w", destPath, cerr)
	}

	// 4. Return sanitized web path
	webPath := "/" + filepath.ToSlash(filepath.Join(DocumentsDir, cleanSubDir, uniqueName))

	return webPath, nil
}

// DeleteStoredFile removes a previously saved file on disk, ignoring missing files.
func DeleteStoredFile(filePath string) error {
	if filePath == "" {
		return nil
	}

	// Sanitise input path
	cleanPath := filepath.Clean(filepath.FromSlash(filePath))
	if strings.HasPrefix(cleanPath, "..") || filepath.IsAbs(cleanPath) {
		// Prevent deleting files outside the workspace
		// Actually, our webPaths start with /documents, so we should be careful
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Ensure we only delete from documents directory
	fullPath := filepath.Join(cwd, cleanPath)
	if !strings.HasPrefix(filepath.ToSlash(fullPath), filepath.ToSlash(filepath.Join(cwd, DocumentsDir))) {
		return fmt.Errorf("security: attempt to delete file outside documents directory")
	}

	slog.Debug("Deleting file", "fullPath", fullPath)

	err = os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return nil
}
