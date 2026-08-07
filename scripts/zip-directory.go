package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: zip-directory <source-directory> <output.zip>")
		os.Exit(2)
	}
	if err := zipDirectory(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func zipDirectory(sourceDirectory, outputPath string) (result error) {
	archive, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create ZIP archive: %w", err)
	}
	defer func() {
		if closeErr := archive.Close(); result == nil && closeErr != nil {
			result = fmt.Errorf("close ZIP archive: %w", closeErr)
		}
	}()

	writer := zip.NewWriter(archive)
	defer func() {
		if closeErr := writer.Close(); result == nil && closeErr != nil {
			result = fmt.Errorf("finalize ZIP archive: %w", closeErr)
		}
	}()

	parentDirectory := filepath.Dir(sourceDirectory)
	return filepath.WalkDir(sourceDirectory, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(parentDirectory, path)
		if err != nil {
			return fmt.Errorf("resolve ZIP path: %w", err)
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("read %s metadata: %w", relativePath, err)
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("create ZIP header: %w", err)
		}
		header.Name = filepath.ToSlash(relativePath)
		header.Method = zip.Deflate
		destination, err := writer.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("add %s to ZIP archive: %w", relativePath, err)
		}
		source, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open %s: %w", relativePath, err)
		}
		_, copyErr := io.Copy(destination, source)
		closeErr := source.Close()
		if copyErr != nil {
			return fmt.Errorf("compress %s: %w", relativePath, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", relativePath, closeErr)
		}
		return nil
	})
}
