package main

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestZipDirectoryPreservesTopLevelDirectory(t *testing.T) {
	temporaryDirectory := t.TempDir()
	sourceDirectory := filepath.Join(temporaryDirectory, "draftmeld_0.3.0_windows_amd64")
	if err := os.MkdirAll(sourceDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDirectory, "draftmeld.exe"), []byte("binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	archivePath := filepath.Join(temporaryDirectory, "draftmeld.zip")
	if err := zipDirectory(sourceDirectory, archivePath); err != nil {
		t.Fatal(err)
	}

	archive, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()

	if len(archive.File) != 1 {
		t.Fatalf("expected one archived file, received %d", len(archive.File))
	}
	if got, want := archive.File[0].Name, "draftmeld_0.3.0_windows_amd64/draftmeld.exe"; got != want {
		t.Fatalf("archive path = %q, want %q", got, want)
	}
}
