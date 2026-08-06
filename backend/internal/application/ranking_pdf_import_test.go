package application

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/document"
)

func TestESPNOverallPDFParserNormalizesSupportedPlayersAndDeduplicatesRanks(t *testing.T) {
	var text strings.Builder
	text.WriteString("2026 ESPN Fantasy Football Draft Kit PPR Top 300 Cheat Sheet\n")
	positions := []string{"RB", "WR", "QB", "TE", "K", "DST"}
	for rank := 1; rank <= 30; rank++ {
		_, _ = fmt.Fprintf(&text, "%d. (%s%d) Player %d, AAA $10 7\n", rank, positions[(rank-1)%len(positions)], rank, rank)
	}
	text.WriteString("1. (RB1) Player 1, AAA $10 7\n")
	parser := espnOverallPDFParser{sourceID: "espn-ppr-pdf", marker: "PPR Top 300 Cheat Sheet"}
	records, published, err := parser.Parse(document.TextDocument{Text: text.String(), PageCount: 1})
	if err != nil {
		t.Fatalf("parse ESPN PDF text: %v", err)
	}
	if len(records) != 30 || records[0].Name != "Player 1" || records[29].Rank != 30 {
		t.Fatalf("unexpected normalized records: %#v", records)
	}
	if published != "User-supplied PDF" {
		t.Fatalf("unexpected publication label: %s", published)
	}
}

func TestLocalESPNPDFAdapters(t *testing.T) {
	directory := os.Getenv("DRAFTMELD_TEST_ESPN_PDF_DIR")
	if directory == "" {
		t.Skip("set DRAFTMELD_TEST_ESPN_PDF_DIR to validate user-supplied ESPN PDFs")
	}
	tests := []struct {
		filename string
		sourceID string
		minimum  int
	}{
		{filename: "ESPN-2026-PPR-Top-300.pdf", sourceID: "espn-ppr-pdf", minimum: 200},
		{filename: "ESPN-2026-Dynasty-Cheat-Sheet.pdf", sourceID: "espn-dynasty-pdf", minimum: 200},
	}
	extractor := document.NativePDFExtractor{}
	for _, test := range tests {
		t.Run(test.sourceID, func(t *testing.T) {
			contents, err := os.ReadFile(filepath.Join(directory, test.filename))
			if err != nil {
				t.Fatalf("read local test PDF: %v", err)
			}
			extracted, err := extractor.Extract(contents)
			if err != nil {
				t.Fatalf("extract local test PDF: %v", err)
			}
			var matched pdfRankingParser
			for _, parser := range defaultPDFRankingParsers() {
				if parser.Matches(extracted) {
					matched = parser
					break
				}
			}
			if matched == nil {
				t.Fatal("no PDF ranking adapter matched")
			}
			records, _, err := matched.Parse(extracted)
			if err != nil {
				t.Fatalf("parse local test PDF: %v", err)
			}
			if len(records) < test.minimum || records[0].SourceID != test.sourceID {
				t.Fatalf("unexpected local import: source=%s records=%d", records[0].SourceID, len(records))
			}
			t.Logf("validated %d normalized records across %d page(s)", len(records), extracted.PageCount)
		})
	}
}
