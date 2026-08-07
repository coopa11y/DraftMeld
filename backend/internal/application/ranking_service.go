package application

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/document"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

type RankingRepository interface {
	ReplaceRankings(context.Context, ranking.SourceDefinition, []ranking.Record, string, time.Time) error
	RankingRecords(context.Context) ([]ranking.Record, error)
	RankingStatuses(context.Context) (map[string]ranking.SourceStatus, error)
	CustomRankingSources(context.Context) ([]ranking.SourceDefinition, error)
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
	definitions, err := service.definitions(ctx)
	if err != nil {
		return nil, err
	}
	statuses := make([]ranking.SourceStatus, 0, len(definitions))
	for _, source := range definitions {
		status := stored[source.ID]
		status.SourceDefinition = source
		statuses = append(statuses, status)
	}
	return statuses, nil
}

func (service *RankingService) definitions(ctx context.Context) ([]ranking.SourceDefinition, error) {
	custom, err := service.repository.CustomRankingSources(ctx)
	if err != nil {
		return nil, err
	}
	definitions := append([]ranking.SourceDefinition(nil), service.sources...)
	return append(definitions, custom...), nil
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
	observedAt := time.Now().UTC()
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
		records, err = resolveRankingPlayers(ctx, service.repository, records, observedAt)
		if err != nil {
			return nil, err
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
		request.Header.Set("User-Agent", "DraftMeld/0.3 (+https://github.com/coopa11y/DraftMeld)")
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

func (service *RankingService) Consensus(ctx context.Context, preferences map[string]league.RankingSourcePreference, methods ...string) ([]ranking.PlayerRanking, error) {
	records, err := service.repository.RankingRecords(ctx)
	if err != nil {
		return nil, err
	}
	aliases, err := service.identityAliases(ctx)
	if err != nil {
		return nil, err
	}
	definitions, err := service.definitions(ctx)
	if err != nil {
		return nil, err
	}
	if err = validateRankingSourcePreferences(definitions, preferences); err != nil {
		return nil, err
	}
	inputs := buildConsensusInputs(definitions, resolveRankingRecords(records, aliases), preferences)
	method := requestedConsensusMethod(methods)
	entries, err := ranking.Combine(inputs.sources, inputs.players, method)
	if err != nil {
		return nil, err
	}
	return overlayCanonicalPlayerMetadata(ctx, service.repository, playerRankings(entries, inputs, method))
}

func effectiveSourcePreference(definition ranking.SourceDefinition, preferences map[string]league.RankingSourcePreference) league.RankingSourcePreference {
	if preference, exists := preferences[definition.ID]; exists {
		return preference
	}
	return league.RankingSourcePreference{Weight: definition.DefaultWeight, Enabled: true}
}
