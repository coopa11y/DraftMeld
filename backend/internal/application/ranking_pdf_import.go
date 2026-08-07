package application

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/document"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

var ErrUnsupportedRankingPDF = errors.New("unsupported ranking PDF")

type PDFImportResult struct {
	Source    ranking.SourceStatus `json:"source"`
	PageCount int                  `json:"pageCount"`
}

type pdfRankingParser interface {
	Matches(document.TextDocument) bool
	Parse(document.TextDocument) ([]ranking.Record, string, error)
}

type espnOverallPDFParser struct {
	sourceID string
	marker   string
}

var (
	espnOverallEntryPattern = regexp.MustCompile(`(?:^|\s)([1-9][0-9]{0,2})\.\s+\((QB|RB|WR|TE|K|DST|D/ST|DEF)[0-9]+\)\s+([^,\r\n]{2,80}),\s+(FA|[A-Z]{2,3})(?:\s+\$?[0-9]|\s+[0-9]{4}-)`)
	espnUpdatedPattern      = regexp.MustCompile(`(?i)(?:Last Update|Updated):\s*([^\r\n]+)`)
)

func defaultPDFRankingParsers() []pdfRankingParser {
	return []pdfRankingParser{
		espnOverallPDFParser{sourceID: "espn-ppr-pdf", marker: "PPR Top 300 Cheat Sheet"},
		espnOverallPDFParser{sourceID: "espn-dynasty-pdf", marker: "Dynasty Cheat Sheet"},
	}
}

func (parser espnOverallPDFParser) Matches(input document.TextDocument) bool {
	return strings.Contains(input.Text, "ESPN Fantasy Football Draft Kit") && strings.Contains(input.Text, parser.marker)
}

func (parser espnOverallPDFParser) Parse(input document.TextDocument) ([]ranking.Record, string, error) {
	matches := espnOverallEntryPattern.FindAllStringSubmatch(input.Text, -1)
	byRank := make(map[int]ranking.Record, len(matches))
	for _, match := range matches {
		rank, err := strconv.Atoi(match[1])
		if err != nil {
			continue
		}
		if _, exists := byRank[rank]; exists {
			continue
		}
		name := strings.Join(strings.Fields(match[3]), " ")
		byRank[rank] = canonicalizeRankingRecord(ranking.Record{SourceID: parser.sourceID, Name: name, Position: match[2], Team: match[4], Rank: rank})
	}
	records := make([]ranking.Record, 0, len(byRank))
	for _, record := range byRank {
		records = append(records, record)
	}
	sort.Slice(records, func(left, right int) bool { return records[left].Rank < records[right].Rank })
	if len(records) < 25 {
		return nil, "", fmt.Errorf("parse ESPN rankings: found only %d usable players", len(records))
	}
	published := "User-supplied PDF"
	if updated := espnUpdatedPattern.FindStringSubmatch(input.Text); len(updated) == 2 {
		published = strings.TrimSpace(updated[1])
	}
	return records, published, nil
}

func (service *RankingService) ImportPDF(ctx context.Context, contents []byte) (PDFImportResult, error) {
	extracted, err := service.pdfExtractor.Extract(contents)
	if err != nil {
		return PDFImportResult{}, err
	}
	for _, parser := range service.pdfParsers {
		if !parser.Matches(extracted) {
			continue
		}
		records, published, parseErr := parser.Parse(extracted)
		if parseErr != nil {
			return PDFImportResult{}, parseErr
		}
		source, exists := service.sourceByID(records[0].SourceID)
		if !exists || source.ImportMode != "pdf-upload" {
			return PDFImportResult{}, fmt.Errorf("PDF parser returned an unknown source")
		}
		records, err = resolveRankingPlayers(ctx, service.repository, records)
		if err != nil {
			return PDFImportResult{}, err
		}
		refreshed := time.Now().UTC()
		if err = service.repository.ReplaceRankings(ctx, source, records, published, refreshed); err != nil {
			return PDFImportResult{}, err
		}
		return PDFImportResult{Source: ranking.SourceStatus{SourceDefinition: source, RecordCount: len(records), RefreshedAt: &refreshed, PublishedAt: published}, PageCount: extracted.PageCount}, nil
	}
	return PDFImportResult{}, fmt.Errorf("%w: upload an ESPN PPR Top 300 or ESPN Dynasty Cheat Sheet; projection and positional-only PDFs are not ranking imports", ErrUnsupportedRankingPDF)
}
