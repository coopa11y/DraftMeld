package application

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

var (
	ErrPlayerUnavailable = errors.New("player is no longer available")
	ErrNothingToUndo     = errors.New("there is no draft action to undo")
)

type DraftEventRepository interface {
	List(context.Context, string) ([]draft.Event, error)
	Append(context.Context, draft.Event) (draft.Event, error)
}

type DraftService struct {
	repository DraftEventRepository
	players    []draft.Player
	playerByID map[string]draft.Player
	leagues    LeagueConfigurationRepository
	mu         sync.Mutex
}

func NewDraftService(
	repository DraftEventRepository,
	players []draft.Player,
	configurations ...LeagueConfiguration,
) (*DraftService, error) {
	if len(configurations) == 0 {
		return nil, errors.New("at least one league configuration is required")
	}
	seenLeagueIDs := make(map[string]struct{}, len(configurations))
	for _, configuration := range configurations {
		if err := configuration.Validate(); err != nil {
			return nil, err
		}
		if _, exists := seenLeagueIDs[configuration.ID]; exists {
			return nil, fmt.Errorf("duplicate league configuration: %s", configuration.ID)
		}
		seenLeagueIDs[configuration.ID] = struct{}{}
	}
	return NewDraftServiceWithLeagues(repository, NewMemoryLeagueRepository(configurations...), players)
}

func NewDraftServiceWithLeagues(
	repository DraftEventRepository,
	leagues LeagueConfigurationRepository,
	players []draft.Player,
) (*DraftService, error) {
	playerByID := make(map[string]draft.Player, len(players))
	for _, player := range players {
		if player.ID == "" {
			return nil, errors.New("draft players require an ID")
		}
		if _, exists := playerByID[player.ID]; exists {
			return nil, fmt.Errorf("duplicate draft player ID: %s", player.ID)
		}
		playerByID[player.ID] = player
	}
	if leagues == nil {
		return nil, errors.New("league configuration repository is required")
	}
	return &DraftService{
		repository: repository, players: players, playerByID: playerByID, leagues: leagues,
	}, nil
}

func (service *DraftService) Snapshot(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, fmt.Errorf("list draft events: %w", err)
	}
	return service.buildSnapshot(configuration, events), nil
}

func (service *DraftService) Record(ctx context.Context, leagueID, playerID string, action draft.Action) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	if _, err := service.configuration(ctx, leagueID); err != nil {
		return draft.Snapshot{}, err
	}
	if action != draft.ActionDraft && action != draft.ActionTaken {
		return draft.Snapshot{}, fmt.Errorf("unsupported draft action: %q", action)
	}
	if _, exists := service.playerByID[playerID]; !exists {
		return draft.Snapshot{}, fmt.Errorf("unknown player: %s", playerID)
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	state := replay(events)
	if _, unavailable := state.playerActions[playerID]; unavailable {
		return draft.Snapshot{}, ErrPlayerUnavailable
	}
	if _, err = service.repository.Append(ctx, draft.Event{LeagueID: leagueID, PlayerID: playerID, Action: action}); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) Undo(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	if _, err := service.configuration(ctx, leagueID); err != nil {
		return draft.Snapshot{}, err
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	state := replay(events)
	if len(state.activeEvents) == 0 {
		return draft.Snapshot{}, ErrNothingToUndo
	}
	target := state.activeEvents[len(state.activeEvents)-1]
	if _, err = service.repository.Append(ctx, draft.Event{
		LeagueID: leagueID, PlayerID: target.PlayerID, Action: draft.ActionUndo, TargetEventID: &target.ID,
	}); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) configuration(ctx context.Context, leagueID string) (LeagueConfiguration, error) {
	configuration, exists, err := service.leagues.GetLeague(ctx, leagueID)
	if err != nil {
		return LeagueConfiguration{}, fmt.Errorf("load league configuration: %w", err)
	}
	if !exists {
		return LeagueConfiguration{}, fmt.Errorf("%w: %s", ErrLeagueNotFound, leagueID)
	}
	return configuration, nil
}

func (service *DraftService) buildSnapshot(configuration LeagueConfiguration, events []draft.Event) draft.Snapshot {
	state := replay(events)
	available := make([]draft.Player, 0, len(service.players))
	myTeam := make([]draft.Player, 0)
	history := make([]draft.Pick, 0, len(state.activeEvents))

	for _, player := range service.players {
		action, unavailable := state.playerActions[player.ID]
		if !unavailable {
			available = append(available, player)
		} else if action == draft.ActionDraft {
			myTeam = append(myTeam, player)
		}
	}
	for index, event := range state.activeEvents {
		history = append(history, draft.Pick{
			EventID: event.ID, Number: index + 1, Action: event.Action,
			Player: service.playerByID[event.PlayerID], CreatedAt: event.CreatedAt,
		})
	}

	return draft.Snapshot{
		LeagueID: configuration.ID, LeagueName: configuration.Rules.Name, PickNumber: len(state.activeEvents) + 1,
		Available: available, MyTeam: myTeam, History: history,
		Recommendations: recommend(available, myTeam, configuration.Rules, configuration.Recommendation),
		CanUndo:         len(state.activeEvents) > 0,
	}
}
