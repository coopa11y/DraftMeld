package document

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	readerpdf "github.com/ledongthuc/pdf"
)

var ErrNoSelectablePDFText = errors.New("no selectable text found")

const (
	MaxPDFBytes  = 20 << 20
	maxPDFPages  = 200
	maxTextBytes = 10 << 20
)

type TextDocument struct {
	Text       string
	PageCount  int
	OCRApplied bool
}

type PDFExtractor interface {
	Extract([]byte) (TextDocument, error)
}

type NativePDFExtractor struct{}

func (NativePDFExtractor) Extract(contents []byte) (result TextDocument, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			result = TextDocument{}
			err = fmt.Errorf("read PDF: invalid document structure")
		}
	}()
	if len(contents) == 0 || len(contents) > MaxPDFBytes {
		return TextDocument{}, fmt.Errorf("read PDF: file must be between 1 byte and 20 MiB")
	}
	if !bytes.HasPrefix(contents, []byte("%PDF-")) {
		return TextDocument{}, fmt.Errorf("read PDF: file does not have a PDF signature")
	}
	reader, err := readerpdf.NewReader(bytes.NewReader(contents), int64(len(contents)))
	if err != nil {
		return TextDocument{}, fmt.Errorf("read PDF: %w", err)
	}
	pageCount := reader.NumPage()
	if pageCount < 1 || pageCount > maxPDFPages {
		return TextDocument{}, fmt.Errorf("read PDF: document must contain 1 to %d pages", maxPDFPages)
	}
	plainText, err := reader.GetPlainText()
	if err != nil {
		return TextDocument{}, fmt.Errorf("extract PDF text: %w", err)
	}
	text, err := io.ReadAll(io.LimitReader(plainText, maxTextBytes+1))
	if err != nil {
		return TextDocument{}, fmt.Errorf("extract PDF text: %w", err)
	}
	if len(text) > maxTextBytes {
		return TextDocument{}, fmt.Errorf("extract PDF text: extracted text exceeds 10 MiB")
	}
	if strings.TrimSpace(string(text)) == "" {
		return TextDocument{}, fmt.Errorf("extract PDF text: %w", ErrNoSelectablePDFText)
	}
	return TextDocument{Text: string(text), PageCount: pageCount}, nil
}
