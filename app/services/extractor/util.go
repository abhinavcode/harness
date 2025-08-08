package extractor

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ZipExtractor provides utilities for extracting zip archives.
type ZipExtractor struct {
	maxFileSize int64
}

// New creates a new ZipExtractor.
func New(maxFileSize int64) *ZipExtractor {
	return &ZipExtractor{
		maxFileSize: maxFileSize,
	}
}

// ExtractZip extracts a zip archive to the specified destination directory.
func (e *ZipExtractor) ExtractZip(zipPath, destDir string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Extract each file
	for _, file := range reader.File {
		destPath := filepath.Join(destDir, file.Name)

		// Create directory if needed
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, file.Mode()); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Create file
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Open source file in the archive
		srcFile, err := file.Open()
		if err != nil {
			destFile.Close()
			return fmt.Errorf("failed to open file in archive: %w", err)
		}

		// Copy file contents
		_, err = io.Copy(destFile, srcFile)

		// Close files
		srcFile.Close()
		destFile.Close()

		if err != nil {
			return fmt.Errorf("failed to extract file: %w", err)
		}
	}

	return nil
}

// ExtractSingleFile extracts a single file from a zip archive.
func (e *ZipExtractor) ExtractSingleFile(zipPath, fileName, destPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Find the file in the archive
	var targetFile *zip.File
	for _, file := range reader.File {
		if file.Name == fileName {
			targetFile = file
			break
		}
	}

	if targetFile == nil {
		return fmt.Errorf("file not found in archive: %s", fileName)
	}

	// Create destination directory if needed
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, targetFile.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Open source file in the archive
	srcFile, err := targetFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in archive: %w", err)
	}
	defer srcFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to extract file: %w", err)
	}

	return nil
}
