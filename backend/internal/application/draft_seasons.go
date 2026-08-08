package application

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

func (service *DraftService) AdvanceSeason(ctx context.Context, leagueID string, season int, draftType league.DraftType, draftOrder []int) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if configuration.Rules.LeagueFormat != league.LeagueFormatDynasty {
		return draft.Snapshot{}, errors.New("season rollover is only available for dynasty leagues")
	}
	if season != configuration.Rules.Season+1 {
		return draft.Snapshot{}, fmt.Errorf("the next season must be %d", configuration.Rules.Season+1)
	}
	if len(draftOrder) != configuration.Rules.TeamCount {
		return draft.Snapshot{}, errors.New("the new draft order must include every franchise")
	}
	existing, err := service.listDraftEvents(ctx, leagueID, season)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if len(existing) > 0 {
		return draft.Snapshot{}, errors.New("the target season already has draft activity")
	}
	configuration.Rules.Season = season
	configuration.Rules.DraftType = draftType
	configuration.Rules.DraftOrder = slices.Clone(draftOrder)
	configuration.Rules.DraftPosition = draftPositionForTeam(draftOrder, userTeamNumber(configuration.Rules))
	if err = configuration.Validate(); err != nil {
		return draft.Snapshot{}, fmt.Errorf("invalid next-season setup: %w", err)
	}
	if err = service.leagues.SaveLeague(ctx, configuration); err != nil {
		return draft.Snapshot{}, fmt.Errorf("save next season: %w", err)
	}
	return service.Snapshot(ctx, leagueID)
}
