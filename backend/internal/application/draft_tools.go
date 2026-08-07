package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

func (service *DraftService) MockToNextTurn(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	currentPick := len(replay(events).activeEvents) + 1
	if isUserTurn(currentPick, configuration.Rules) {
		return draft.Snapshot{}, errors.New("make your pick before simulating opponent selections")
	}
	targetPick := nextUserPick(currentPick, configuration.Rules)
	if targetPick == 0 || targetPick <= currentPick {
		return draft.Snapshot{}, errors.New("unable to determine the next draft turn")
	}
	players, _, _, _, err := service.playersForLeague(ctx, configuration)
	if err != nil {
		return draft.Snapshot{}, err
	}
	unavailable := replay(events).playerActions
	available := make([]draft.Player, 0, len(players))
	for _, player := range players {
		if _, taken := unavailable[player.ID]; !taken {
			available = append(available, player)
		}
	}
	for pick := currentPick; pick < targetPick && len(available) > 0; pick++ {
		index := mockSelection(available, pick)
		selected := available[index]
		if _, err = service.repository.Append(ctx, draft.Event{LeagueID: leagueID, PlayerID: selected.ID, Action: draft.ActionTaken}); err != nil {
			return draft.Snapshot{}, err
		}
		available = append(available[:index], available[index+1:]...)
	}
	return service.Snapshot(ctx, leagueID)
}

func mockSelection(players []draft.Player, pick int) int {
	order := make([]int, len(players))
	for index := range players {
		order[index] = index
	}
	positionBias := []string{"RB", "WR", "WR", "RB", "QB", "TE"}[pick%6]
	sort.SliceStable(order, func(left, right int) bool {
		leftPlayer, rightPlayer := players[order[left]], players[order[right]]
		leftScore, rightScore := leftPlayer.ADP, rightPlayer.ADP
		if leftPlayer.Position == positionBias {
			leftScore -= 2.5
		}
		if rightPlayer.Position == positionBias {
			rightScore -= 2.5
		}
		return leftScore < rightScore
	})
	choice := (pick*7 + 3) % min(4, len(order))
	return order[choice]
}

type sleeperPick struct {
	PickNumber int `json:"pick_no"`
	RosterID   int `json:"roster_id"`
	Metadata   struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Position  string `json:"position"`
		Team      string `json:"team"`
	} `json:"metadata"`
}

func (service *DraftService) SyncSleeper(ctx context.Context, leagueID, sleeperDraftID string, myRosterID int) (draft.Snapshot, error) {
	if strings.TrimSpace(sleeperDraftID) == "" || myRosterID < 1 {
		return draft.Snapshot{}, errors.New("Sleeper draft ID and your roster ID are required")
	}
	endpoint := strings.TrimRight(service.sleeperBaseURL, "/") + "/draft/" + url.PathEscape(strings.TrimSpace(sleeperDraftID)) + "/picks"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return draft.Snapshot{}, err
	}
	response, err := service.sleeperClient.Do(request)
	if err != nil {
		return draft.Snapshot{}, fmt.Errorf("load Sleeper picks: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return draft.Snapshot{}, fmt.Errorf("Sleeper returned HTTP %d", response.StatusCode)
	}
	var picks []sleeperPick
	if err = json.NewDecoder(response.Body).Decode(&picks); err != nil {
		return draft.Snapshot{}, errors.New("Sleeper returned an invalid draft response")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	players, playerByID, _, _, err := service.playersForLeague(ctx, configuration)
	if err != nil {
		return draft.Snapshot{}, err
	}
	playerIDByIdentity := make(map[string]string, len(players))
	for _, player := range players {
		playerIDByIdentity[canonicalRankingKey(player.Name, player.Position, player.NFLTeam)] = player.ID
	}
	events, err := service.repository.List(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	state := replay(events)
	sort.Slice(picks, func(left, right int) bool { return picks[left].PickNumber < picks[right].PickNumber })
	for _, pick := range picks {
		name := strings.TrimSpace(pick.Metadata.FirstName + " " + pick.Metadata.LastName)
		playerID := playerIDByIdentity[canonicalRankingKey(name, pick.Metadata.Position, pick.Metadata.Team)]
		if _, known := playerByID[playerID]; !known {
			continue
		}
		if _, exists := state.playerActions[playerID]; exists {
			continue
		}
		action := draft.ActionTaken
		if pick.RosterID == myRosterID {
			action = draft.ActionDraft
		}
		if _, err = service.repository.Append(ctx, draft.Event{LeagueID: leagueID, PlayerID: playerID, Action: action}); err != nil {
			return draft.Snapshot{}, err
		}
		state.playerActions[playerID] = action
	}
	return service.Snapshot(ctx, leagueID)
}
