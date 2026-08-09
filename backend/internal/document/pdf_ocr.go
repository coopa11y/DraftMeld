package document

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	readerpdf "github.com/ledongthuc/pdf"
)

const (
	maxOCRPages       = 25
	maxOCRImageBytes  = 50 << 20
	maxOCRTotalBytes  = 200 << 20
	maxOCRStderrBytes = 4 << 10
	ocrTimeout        = 2 * time.Minute
)

var ErrOCRUnavailable = errors.New("local PDF OCR is unavailable")
var ocrSlots = make(chan struct{}, 1)

type AutoPDFExtractor struct {
	selectable PDFExtractor
	ocr        PDFExtractor
}

func NewPDFExtractor() PDFExtractor {
	return AutoPDFExtractor{selectable: NativePDFExtractor{}, ocr: NewCommandPDFOCRExtractor()}
}

func (extractor AutoPDFExtractor) Extract(contents []byte) (TextDocument, error) {
	result, err := extractor.selectable.Extract(contents)
	if err == nil || !errors.Is(err, ErrNoSelectablePDFText) {
		return result, err
	}
	return extractor.ocr.Extract(contents)
}

type commandPDFOCRExtractor struct {
	inspect  func([]byte) (int, error)
	lookPath func(string) (string, error)
	run      func(context.Context, string, ...string) ([]byte, error)
	tempDir  func(string, string) (string, error)
}

func NewCommandPDFOCRExtractor() PDFExtractor {
	return commandPDFOCRExtractor{
		inspect:  inspectOCRPDF,
		lookPath: exec.LookPath,
		run:      runOCRCommand,
		tempDir:  os.MkdirTemp,
	}
}

func (extractor commandPDFOCRExtractor) Extract(contents []byte) (TextDocument, error) {
	pageCount, err := extractor.inspect(contents)
	if err != nil {
		return TextDocument{}, err
	}
	renderer, err := extractor.resolve("DRAFTMELD_PDFTOPPM_PATH", "pdftoppm")
	if err != nil {
		return TextDocument{}, err
	}
	ocrEngine, err := extractor.resolve("DRAFTMELD_TESSERACT_PATH", "tesseract")
	if err != nil {
		return TextDocument{}, err
	}
	contextWithTimeout, cancel := context.WithTimeout(context.Background(), ocrTimeout)
	defer cancel()
	select {
	case ocrSlots <- struct{}{}:
		defer func() { <-ocrSlots }()
	case <-contextWithTimeout.Done():
		return TextDocument{}, errors.New("PDF OCR is busy; try again in a moment")
	}
	directory, err := extractor.tempDir("", "draftmeld-ocr-*")
	if err != nil {
		return TextDocument{}, fmt.Errorf("prepare PDF OCR: %w", err)
	}
	defer os.RemoveAll(directory)
	pdfPath := filepath.Join(directory, "document.pdf")
	if err = os.WriteFile(pdfPath, contents, 0o600); err != nil {
		return TextDocument{}, fmt.Errorf("prepare PDF OCR: %w", err)
	}

	prefix := filepath.Join(directory, "page")
	if output, runErr := extractor.run(contextWithTimeout, renderer, "-png", "-r", "180", "-f", "1", "-l", strconv.Itoa(pageCount), pdfPath, prefix); runErr != nil {
		return TextDocument{}, commandError("render scanned PDF", runErr, output, contextWithTimeout.Err())
	}
	images, err := filepath.Glob(prefix + "-*.png")
	if err != nil || len(images) != pageCount {
		return TextDocument{}, fmt.Errorf("render scanned PDF: expected %d page images, found %d", pageCount, len(images))
	}
	sort.Slice(images, func(left, right int) bool { return pageNumber(images[left]) < pageNumber(images[right]) })

	var text strings.Builder
	totalImageBytes := int64(0)
	for _, image := range images {
		info, statErr := os.Stat(image)
		if statErr != nil {
			return TextDocument{}, fmt.Errorf("inspect OCR page: %w", statErr)
		}
		if info.Size() > maxOCRImageBytes || totalImageBytes+info.Size() > maxOCRTotalBytes {
			return TextDocument{}, errors.New("OCR page images exceed the safe processing limit")
		}
		totalImageBytes += info.Size()
		output, runErr := extractor.run(contextWithTimeout, ocrEngine, image, "stdout", "-l", "eng", "--oem", "1")
		if runErr != nil {
			return TextDocument{}, commandError("recognize scanned PDF text", runErr, output, contextWithTimeout.Err())
		}
		if text.Len()+len(output) > maxTextBytes {
			return TextDocument{}, errors.New("OCR text exceeds the 10 MiB processing limit")
		}
		text.Write(output)
		text.WriteByte('\n')
	}
	if strings.TrimSpace(text.String()) == "" {
		return TextDocument{}, errors.New("OCR completed but no readable English text was found")
	}
	return TextDocument{Text: text.String(), PageCount: pageCount, OCRApplied: true}, nil
}

func (extractor commandPDFOCRExtractor) resolve(environmentName, fallback string) (string, error) {
	configured := strings.TrimSpace(os.Getenv(environmentName))
	if configured == "" {
		configured = fallback
	}
	path, err := extractor.lookPath(configured)
	if err != nil {
		return "", fmt.Errorf("%w: %s was not found; install Poppler and Tesseract or use the DraftMeld container", ErrOCRUnavailable, configured)
	}
	return path, nil
}

func inspectOCRPDF(contents []byte) (int, error) {
	if len(contents) == 0 || len(contents) > MaxPDFBytes {
		return 0, errors.New("read PDF: file must be between 1 byte and 20 MiB")
	}
	if !bytes.HasPrefix(contents, []byte("%PDF-")) {
		return 0, errors.New("read PDF: file does not have a PDF signature")
	}
	reader, err := readerpdf.NewReader(bytes.NewReader(contents), int64(len(contents)))
	if err != nil {
		return 0, fmt.Errorf("read PDF for OCR: %w", err)
	}
	pageCount := reader.NumPage()
	if pageCount < 1 || pageCount > maxOCRPages {
		return 0, fmt.Errorf("OCR supports scanned PDFs containing 1 to %d pages", maxOCRPages)
	}
	return pageCount, nil
}

func runOCRCommand(ctx context.Context, name string, arguments ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, name, arguments...)
	stdout := newBoundedBuffer(maxTextBytes + 1)
	stderr := newBoundedBuffer(maxOCRStderrBytes)
	command.Stdout = stdout
	command.Stderr = stderr
	if err := command.Run(); err != nil {
		return stderr.Bytes(), err
	}
	return stdout.Bytes(), nil
}

type boundedBuffer struct {
	mutex    sync.Mutex
	contents bytes.Buffer
	limit    int
}

func newBoundedBuffer(limit int) *boundedBuffer {
	return &boundedBuffer{limit: limit}
}

func (buffer *boundedBuffer) Write(value []byte) (int, error) {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	remaining := buffer.limit - buffer.contents.Len()
	if remaining > 0 {
		if remaining < len(value) {
			_, _ = buffer.contents.Write(value[:remaining])
		} else {
			_, _ = buffer.contents.Write(value)
		}
	}
	return len(value), nil
}

func (buffer *boundedBuffer) Bytes() []byte {
	buffer.mutex.Lock()
	defer buffer.mutex.Unlock()
	return bytes.Clone(buffer.contents.Bytes())
}

func commandError(action string, err error, output []byte, contextErr error) error {
	if errors.Is(contextErr, context.DeadlineExceeded) {
		return fmt.Errorf("%s: OCR exceeded the two-minute processing limit", action)
	}
	detail := strings.TrimSpace(string(output))
	if len(detail) > 500 {
		detail = detail[:500]
	}
	if detail == "" {
		return fmt.Errorf("%s: %w", action, err)
	}
	return fmt.Errorf("%s: %w: %s", action, err, detail)
}

func pageNumber(path string) int {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	value, _ := strconv.Atoi(base[strings.LastIndex(base, "-")+1:])
	return value
}
