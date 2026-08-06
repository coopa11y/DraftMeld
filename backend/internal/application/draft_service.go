package application

import (
	"context"
	"errors"
	"fmt"
	"sort"
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
	mu         sync.Mutex
}

func NewDraftService(repository DraftEventRepository, players []draft.Player) *DraftService {
	playerByID := make(map[string]draft.Player, len(players))
	for _, player := range players {
		playerByID[player.ID] = player
	}
	return &DraftService{repository: repository, players: players, playerByID: playerByID}
}

func (service *DraftService) Snapshot(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, fmt.Errorf("list draft events: %w", err)
	}
	return service.buildSnapshot(leagueID, events), nil
}

func (service *DraftService) Record(ctx context.Context, leagueID, playerID string, action draft.Action) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

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
	states, _, _ := replay(events)
	if _, unavailable := states[playerID]; unavailable {
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

	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	_, active, _ := replay(events)
	if len(active) == 0 {
		return draft.Snapshot{}, ErrNothingToUndo
	}
	target := active[len(active)-1]
	if _, err = service.repository.Append(ctx, draft.Event{
		LeagueID: leagueID, PlayerID: target.PlayerID, Action: draft.ActionUndo, TargetEventID: &target.ID,
	}); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) buildSnapshot(leagueID string, events []draft.Event) draft.Snapshot {
	states, active, _ := replay(events)
	available := make([]draft.Player, 0, len(service.players))
	myTeam := make([]draft.Player, 0)
	history := make([]draft.Pick, 0, len(active))

	for _, player := range service.players {
		action, unavailable := states[player.ID]
		if !unavailable {
			available = append(available, player)
		} else if action == draft.ActionDraft {
			myTeam = append(myTeam, player)
		}
	}
	for index, event := range active {
		history = append(history, draft.Pick{
			EventID: event.ID, Number: index + 1, Action: event.Action,
			Player: service.playerByID[event.PlayerID], CreatedAt: event.CreatedAt,
		})
	}

	return draft.Snapshot{
		LeagueID: leagueID, LeagueName: "Demo League", PickNumber: len(active) + 1,
		Available: available, MyTeam: myTeam, History: history,
		Recommendations: recommend(available, myTeam), CanUndo: len(active) > 0,
	}
}

func replay(events []draft.Event) (map[string]draft.Action, []draft.Event, map[int64]bool) {
	undone := make(map[int64]bool)
	for _, event := range events {
		if event.Action == draft.ActionUndo && event.TargetEventID != nil {
			undone[*event.TargetEventID] = true
		}
	}
	states := make(map[string]draft.Action)
	active := make([]draft.Event, 0)
	for _, event := range events {
		if event.Action == draft.ActionUndo || undone[event.ID] {
			continue
		}
		states[event.PlayerID] = event.Action
		active = append(active, event)
	}
	return states, active, undone
}

func recommend(available, myTeam []draft.Player) []draft.Recommendation {
	needs := map[string]int{"QB": 1, "RB": 2, "WR": 2, "TE": 1}
	for _, player := range myTeam {
		if needs[player.Position] > 0 {
			needs[player.Position]--
		}
	}
	positionRanks := make(map[string][]int)
	for _, player := range available {
		positionRanks[player.Position] = append(positionRanks[player.Position], player.OverallRank)
	}

	recommendations := make([]draft.Recommendation, 0, len(available))
	for _, player := range available {
		score := 200.0 - float64(player.OverallRank)
		reasons := make([]string, 0, 3)
		if needs[player.Position] > 0 {
			score += 24
			reasons = append(reasons, "Fills an open starting roster need")
		}
		value := player.ADP - float64(player.OverallRank)
		if value >= 5 {
			score += value
			reasons = append(reasons, fmt.Sprintf("Ranks %.0f spots above draft-room ADP", value))
		}
		ranks := positionRanks[player.Position]
		if len(ranks) > 1 && ranks[0] == player.OverallRank && ranks[1]-ranks[0] >= 5 {
			score += 8
			reasons = append(reasons, "Top option before a positional drop-off")
		}
		if len(reasons) == 0 {
			reasons = append(reasons, "Best available league-adjusted value")
		}
		recommendations = append(recommendations, draft.Recommendation{Player: player, Score: score, Reasons: reasons})
	}
	sort.SliceStable(recommendations, func(left, right int) bool {
		return recommendations[left].Score > recommendations[right].Score
	})
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}
	return recommendations
}
