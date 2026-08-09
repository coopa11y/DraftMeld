package document

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type pdfExtractorStub struct {
	document TextDocument
	err      error
}

func (stub pdfExtractorStub) Extract([]byte) (TextDocument, error) {
	return stub.document, stub.err
}

func TestAutoPDFExtractorUsesOCRForScannedDocuments(t *testing.T) {
	expected := TextDocument{Text: "League Name: Scanned League", PageCount: 1, OCRApplied: true}
	extractor := AutoPDFExtractor{
		selectable: pdfExtractorStub{err: fmtNoSelectableText()},
		ocr:        pdfExtractorStub{document: expected},
	}
	result, err := extractor.Extract([]byte("pdf"))
	if err != nil || result != expected {
		t.Fatalf("unexpected OCR fallback result: %#v, %v", result, err)
	}
}

func TestAutoPDFExtractorDoesNotHideInvalidPDFErrors(t *testing.T) {
	invalid := errors.New("invalid PDF")
	extractor := AutoPDFExtractor{
		selectable: pdfExtractorStub{err: invalid},
		ocr:        pdfExtractorStub{document: TextDocument{OCRApplied: true}},
	}
	_, err := extractor.Extract([]byte("pdf"))
	if !errors.Is(err, invalid) {
		t.Fatalf("expected original validation error, got %v", err)
	}
}

func TestCommandPDFOCRExtractorRendersAndRecognizesPagesInOrder(t *testing.T) {
	root := t.TempDir()
	temporary := filepath.Join(root, "ocr")
	ocrCalls := 0
	extractor := commandPDFOCRExtractor{
		inspect:  func([]byte) (int, error) { return 2, nil },
		lookPath: func(name string) (string, error) { return name, nil },
		tempDir: func(string, string) (string, error) {
			return temporary, os.Mkdir(temporary, 0o700)
		},
		run: func(_ context.Context, name string, arguments ...string) ([]byte, error) {
			if name == "pdftoppm" {
				prefix := arguments[len(arguments)-1]
				if err := os.WriteFile(prefix+"-2.png", []byte("page two"), 0o600); err != nil {
					return nil, err
				}
				if err := os.WriteFile(prefix+"-1.png", []byte("page one"), 0o600); err != nil {
					return nil, err
				}
				return nil, nil
			}
			ocrCalls++
			if strings.HasSuffix(arguments[0], "-1.png") {
				return []byte("League Name: Scanned League\n"), nil
			}
			return []byte("Number of Teams: 10\n"), nil
		},
	}
	result, err := extractor.Extract([]byte("%PDF-fake"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.OCRApplied || result.PageCount != 2 || ocrCalls != 2 || !strings.HasPrefix(result.Text, "League Name") || !strings.Contains(result.Text, "Number of Teams") {
		t.Fatalf("unexpected OCR result: %#v, calls=%d", result, ocrCalls)
	}
	if _, err = os.Stat(temporary); !os.IsNotExist(err) {
		t.Fatalf("temporary OCR directory was not removed: %v", err)
	}
}

func TestCommandPDFOCRExtractorReportsMissingTools(t *testing.T) {
	extractor := commandPDFOCRExtractor{
		inspect:  func([]byte) (int, error) { return 1, nil },
		lookPath: func(name string) (string, error) { return "", exec.ErrNotFound },
	}
	_, err := extractor.Extract([]byte("%PDF-fake"))
	if !errors.Is(err, ErrOCRUnavailable) || !strings.Contains(err.Error(), "use the DraftMeld container") {
		t.Fatalf("expected actionable OCR capability error, got %v", err)
	}
}

func TestSyntheticScannedPDFFixtureRequiresOCR(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join("testdata", "scanned-league-rules.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = (NativePDFExtractor{}).Extract(contents)
	if !errors.Is(err, ErrNoSelectablePDFText) {
		t.Fatalf("expected scanned fixture to contain no selectable text, got %v", err)
	}
}

func TestCommandPDFOCRExtractorWithInstalledTools(t *testing.T) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		t.Skip("pdftoppm is not installed")
	}
	if _, err := exec.LookPath("tesseract"); err != nil {
		t.Skip("tesseract is not installed")
	}
	contents, err := os.ReadFile(filepath.Join("testdata", "scanned-league-rules.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := NewCommandPDFOCRExtractor().Extract(contents)
	if err != nil {
		t.Fatal(err)
	}
	if !result.OCRApplied || !strings.Contains(result.Text, "OCR Test League") || !strings.Contains(result.Text, "Passing Touchdowns") {
		t.Fatalf("unexpected installed-tool OCR output: %#v", result)
	}
}

func TestBoundedBufferDiscardsBytesBeyondItsLimit(t *testing.T) {
	buffer := newBoundedBuffer(5)
	written, err := buffer.Write([]byte("123456789"))
	if err != nil || written != 9 || string(buffer.Bytes()) != "12345" {
		t.Fatalf("unexpected bounded output: %q, written=%d, err=%v", buffer.Bytes(), written, err)
	}
}

func fmtNoSelectableText() error {
	return errors.Join(errors.New("extract PDF text"), ErrNoSelectablePDFText)
}
