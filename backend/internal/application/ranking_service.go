package application

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/document"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type RankingRepository interface {
	ReplaceRankings(context.Context, ranking.SourceDefinition, []ranking.Record, string, time.Time) error
	RankingRecords(context.Context) ([]ranking.Record, error)
	RankingStatuses(context.Context) (map[string]ranking.SourceStatus, error)
}

type RankingService struct {
	repository   RankingRepository
	client       *http.Client
	sources      []ranking.SourceDefinition
	pdfExtractor document.PDFExtractor
	pdfParsers   []pdfRankingParser
}

func NewRankingService(repository RankingRepository) *RankingService {
	return &RankingService{repository: repository, client: &http.Client{Timeout: 60 * time.Second}, sources: BuiltInRankingSources(), pdfExtractor: document.NativePDFExtractor{}, pdfParsers: defaultPDFRankingParsers()}
}

func (service *RankingService) Sources(ctx context.Context) ([]ranking.SourceStatus, error) {
	stored, err := service.repository.RankingStatuses(ctx)
	if err != nil {
		return nil, err
	}
	statuses := make([]ranking.SourceStatus, 0, len(service.sources))
	for _, source := range service.sources {
		status := stored[source.ID]
		status.SourceDefinition = source
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func (service *RankingService) sourceByID(id string) (ranking.SourceDefinition, bool) {
	for _, source := range service.sources {
		if source.ID == id {
			return source, true
		}
	}
	return ranking.SourceDefinition{}, false
}

func (service *RankingService) Refresh(ctx context.Context) ([]ranking.SourceStatus, error) {
	downloads := make(map[string][]byte)
	for _, source := range service.sources {
		if source.ImportMode != "download" {
			continue
		}
		contents, exists := downloads[source.DataURL]
		if !exists {
			var err error
			contents, err = service.download(ctx, source)
			if err != nil {
				return nil, err
			}
			downloads[source.DataURL] = contents
		}
		records, published, err := parseRankingSource(source.ID, bytes.NewReader(contents))
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", source.Name, err)
		}
		if len(records) == 0 {
			return nil, fmt.Errorf("parse %s: no usable players", source.Name)
		}
		if err = service.repository.ReplaceRankings(ctx, source, records, published, time.Now().UTC()); err != nil {
			return nil, err
		}
	}
	return service.Sources(ctx)
}

func (service *RankingService) download(ctx context.Context, source ranking.SourceDefinition) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.DataURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build %s request: %w", source.Name, err)
		}
		request.Header.Set("User-Agent", "DraftMeld/0.2 (+https://github.com/coopa11y/DraftMeld)")
		response, err := service.client.Do(request)
		if err != nil {
			lastErr = err
			continue
		}
		if response.StatusCode != http.StatusOK {
			response.Body.Close()
			if response.StatusCode >= http.StatusInternalServerError {
				lastErr = fmt.Errorf("HTTP %d", response.StatusCode)
				continue
			}
			return nil, fmt.Errorf("download %s: HTTP %d", source.Name, response.StatusCode)
		}
		contents, readErr := io.ReadAll(io.LimitReader(response.Body, (20<<20)+1))
		response.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if len(contents) > 20<<20 {
			return nil, fmt.Errorf("download %s: response exceeds 20 MiB", source.Name)
		}
		return contents, nil
	}
	return nil, fmt.Errorf("download %s after retry: %w", source.Name, lastErr)
}

func (service *RankingService) Consensus(ctx context.Context, requestedWeights map[string]float64) ([]ranking.PlayerRanking, error) {
	records, err := service.repository.RankingRecords(ctx)
	if err != nil {
		return nil, err
	}
	definitions := make(map[string]ranking.SourceDefinition, len(service.sources))
	for _, source := range service.sources {
		definitions[source.ID] = source
	}
	// The current redraft consensus defines the eligible draft pool. Historical
	// opportunity and dynasty feeds enrich those players without introducing
	// retired or otherwise undraftable players on their own.
	eligible := make(map[string]struct{})
	for _, record := range records {
		record = canonicalizeRankingRecord(record)
		if record.SourceID == "redraft-ecr" || record.SourceID == "espn-ppr-pdf" {
			eligible[record.PlayerKey] = struct{}{}
		}
	}
	sources := make(map[string]ranking.Source)
	metadata := make(map[string]ranking.Record)
	sourceRanks := make(map[string]map[string]int)
	for _, record := range records {
		record = canonicalizeRankingRecord(record)
		if _, exists := eligible[record.PlayerKey]; !exists {
			continue
		}
		definition, exists := definitions[record.SourceID]
		if !exists {
			continue
		}
		source := sources[record.SourceID]
		weight := definition.DefaultWeight
		if requested, configured := requestedWeights[record.SourceID]; configured {
			if requested <= 0 || requested > 10 {
				return nil, fmt.Errorf("ranking source %s requires a weight greater than 0 and no more than 10", record.SourceID)
			}
			weight = requested
		}
		source.ID, source.Weight = record.SourceID, weight
		if source.Ranks == nil {
			source.Ranks = make(map[string]int)
		}
		source.Ranks[record.PlayerKey] = record.Rank
		sources[record.SourceID] = source
		if _, exists = metadata[record.PlayerKey]; !exists || record.SourceID == "espn-ppr-pdf" || record.SourceID == "redraft-ecr" {
			metadata[record.PlayerKey] = record
		}
		if sourceRanks[record.PlayerKey] == nil {
			sourceRanks[record.PlayerKey] = make(map[string]int)
		}
		sourceRanks[record.PlayerKey][record.SourceID] = record.Rank
	}
	weighted := make([]ranking.Source, 0, len(sources))
	for _, source := range sources {
		weighted = append(weighted, source)
	}
	entries, err := ranking.WeightedAverage(weighted)
	if err != nil {
		return nil, err
	}
	result := make([]ranking.PlayerRanking, 0, len(entries))
	for index, entry := range entries {
		player := metadata[entry.PlayerID]
		result = append(result, ranking.PlayerRanking{PlayerKey: entry.PlayerID, Name: player.Name, Position: player.Position, Team: player.Team, Rank: index + 1, Score: entry.Score, SourceCount: entry.SourceCount, SourceRanks: sourceRanks[entry.PlayerID]})
	}
	return result, nil
}
