package document

import (
	"strings"
	"testing"
)

func TestNativePDFExtractorRejectsNonPDFContent(t *testing.T) {
	_, err := (NativePDFExtractor{}).Extract([]byte("not a PDF"))
	if err == nil || !strings.Contains(err.Error(), "PDF signature") {
		t.Fatalf("expected PDF signature error, got %v", err)
	}
}

func TestNativePDFExtractorRejectsEmptyContent(t *testing.T) {
	_, err := (NativePDFExtractor{}).Extract(nil)
	if err == nil || !strings.Contains(err.Error(), "between 1 byte and 20 MiB") {
		t.Fatalf("expected size error, got %v", err)
	}
}
