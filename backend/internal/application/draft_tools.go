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
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/player"
)

func (service *DraftService) MockToNextTurn(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
	if err != nil {
		return draft.Snapshot{}, err
	}
	events, err = resolveDraftEventAliases(ctx, service.repository, events)
	if err != nil {
		return draft.Snapshot{}, err
	}
	activeEvents := replay(events).activeEvents
	if err = service.requireDraftStarted(ctx, leagueID, configuration.Rules.Season, len(activeEvents), totalDraftPicks(configuration.Rules)); err != nil {
		return draft.Snapshot{}, err
	}
	currentPick := len(activeEvents) + 1
	trades, err := service.pickTrades(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if ownerForPick(currentPick, configuration.Rules, trades) == userTeamNumber(configuration.Rules) {
		return draft.Snapshot{}, errors.New("make your pick before simulating opponent selections")
	}
	targetPick := nextUserPickWithTrades(currentPick, configuration.Rules, trades)
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
		if _, err = service.repository.Append(ctx, draft.Event{LeagueID: leagueID, Season: configuration.Rules.Season, PlayerID: selected.ID, Action: draft.ActionTaken, TeamNumber: ownerForPick(pick, configuration.Rules, trades)}); err != nil {
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
	PickNumber int    `json:"pick_no"`
	RosterID   int    `json:"roster_id"`
	PlayerID   string `json:"player_id"`
	Metadata   struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Position  string `json:"position"`
		Team      string `json:"team"`
	} `json:"metadata"`
}

type SleeperSyncResult struct {
	Snapshot  draft.Snapshot `json:"snapshot"`
	Added     int            `json:"added"`
	Updated   int            `json:"updated"`
	Removed   int            `json:"removed"`
	Unmatched int            `json:"unmatched"`
}

type DraftReconciliationRepository interface {
	ReplaceDraftEvents(context.Context, string, []draft.Event) error
}

type seasonDraftReconciliationRepository interface {
	ReplaceDraftEventsForSeason(context.Context, string, int, []draft.Event) error
}

func (service *DraftService) SyncSleeper(ctx context.Context, leagueID, sleeperDraftID string, myRosterID int) (SleeperSyncResult, error) {
	if strings.TrimSpace(sleeperDraftID) == "" || myRosterID < 1 {
		return SleeperSyncResult{}, errors.New("Sleeper draft ID and your roster ID are required")
	}
	endpoint := strings.TrimRight(service.sleeperBaseURL, "/") + "/draft/" + url.PathEscape(strings.TrimSpace(sleeperDraftID)) + "/picks"
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return SleeperSyncResult{}, err
	}
	response, err := service.sleeperClient.Do(request)
	if err != nil {
		return SleeperSyncResult{}, fmt.Errorf("load Sleeper picks: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return SleeperSyncResult{}, fmt.Errorf("Sleeper returned HTTP %d", response.StatusCode)
	}
	var picks []sleeperPick
	if err = json.NewDecoder(response.Body).Decode(&picks); err != nil {
		return SleeperSyncResult{}, errors.New("Sleeper returned an invalid draft response")
	}
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return SleeperSyncResult{}, err
	}
	players, playerByID, _, _, err := service.playersForLeague(ctx, configuration)
	if err != nil {
		return SleeperSyncResult{}, err
	}
	playerIDByIdentity := make(map[string]string, len(players))
	for _, player := range players {
		playerIDByIdentity[canonicalRankingKey(player.Name, player.Position, player.NFLTeam)] = player.ID
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
	if err != nil {
		return SleeperSyncResult{}, err
	}
	events, err = resolveDraftEventAliases(ctx, service.repository, events)
	if err != nil {
		return SleeperSyncResult{}, err
	}
	state := replay(events)
	if err = service.requireDraftStarted(ctx, leagueID, configuration.Rules.Season, len(state.activeEvents), totalDraftPicks(configuration.Rules)); err != nil {
		return SleeperSyncResult{}, err
	}
	previous := make(map[string]draft.Action, len(state.playerActions))
	for playerID, action := range state.playerActions {
		previous[playerID] = action
	}
	result := SleeperSyncResult{}
	reconciled := make([]draft.Event, 0, len(picks))
	teamNumbers := sleeperTeamNumbers(picks, myRosterID, configuration.Rules)
	current := make(map[string]draft.Action)
	sort.Slice(picks, func(left, right int) bool { return picks[left].PickNumber < picks[right].PickNumber })
	for _, pick := range picks {
		name := strings.TrimSpace(pick.Metadata.FirstName + " " + pick.Metadata.LastName)
		identityKey := canonicalRankingKey(name, pick.Metadata.Position, pick.Metadata.Team)
		playerID := ""
		if directory, ok := service.repository.(PlayerDirectoryRepository); ok && pick.PlayerID != "" {
			proposedID, idErr := newCanonicalPlayerID()
			if idErr != nil {
				return SleeperSyncResult{}, idErr
			}
			resolved, resolveErr := directory.ResolvePlayer(ctx, player.Candidate{
				IdentityKey: identityKey, LegacyKey: identityKey, Name: name,
				Position: normalizePosition(pick.Metadata.Position), Team: pick.Metadata.Team,
				Provider: "sleeper", ProviderID: pick.PlayerID, ObservedAt: time.Now().UTC(),
			}, proposedID)
			if resolveErr != nil {
				return SleeperSyncResult{}, resolveErr
			}
			if _, known := playerByID[resolved.ID]; known {
				playerID = resolved.ID
			}
		}
		if playerID == "" {
			playerID = playerIDByIdentity[identityKey]
		}
		if _, known := playerByID[playerID]; !known {
			result.Unmatched++
			continue
		}
		if _, exists := current[playerID]; exists {
			continue
		}
		action := draft.ActionTaken
		if pick.RosterID == myRosterID {
			action = draft.ActionDraft
		}
		current[playerID] = action
		reconciled = append(reconciled, draft.Event{LeagueID: leagueID, Season: configuration.Rules.Season, PlayerID: playerID, Action: action, TeamNumber: teamNumbers[pick.RosterID]})
		if prior, exists := previous[playerID]; !exists {
			result.Added++
		} else if prior != action {
			result.Updated++
		}
	}
	for playerID := range previous {
		if _, exists := current[playerID]; !exists {
			result.Removed++
		}
	}
	if len(picks) > 0 && len(reconciled) == 0 {
		return SleeperSyncResult{}, errors.New("Sleeper picks did not contain any players DraftMeld could match; local history was left unchanged")
	}
	if repository, ok := service.repository.(seasonDraftReconciliationRepository); ok {
		if err = repository.ReplaceDraftEventsForSeason(ctx, leagueID, configuration.Rules.Season, reconciled); err != nil {
			return SleeperSyncResult{}, err
		}
	} else if repository, ok := service.repository.(DraftReconciliationRepository); ok {
		if err = repository.ReplaceDraftEvents(ctx, leagueID, reconciled); err != nil {
			return SleeperSyncResult{}, err
		}
	} else {
		for _, event := range reconciled {
			if _, exists := previous[event.PlayerID]; exists {
				continue
			}
			if _, err = service.repository.Append(ctx, event); err != nil {
				return SleeperSyncResult{}, err
			}
		}
	}
	result.Snapshot, err = service.Snapshot(ctx, leagueID)
	return result, err
}

func sleeperTeamNumbers(picks []sleeperPick, myRosterID int, rules league.Rules) map[int]int {
	result := map[int]int{myRosterID: userTeamNumber(rules)}
	rosterIDs := make([]int, 0)
	seen := map[int]bool{myRosterID: true}
	for _, pick := range picks {
		if !seen[pick.RosterID] {
			seen[pick.RosterID] = true
			rosterIDs = append(rosterIDs, pick.RosterID)
		}
	}
	sort.Ints(rosterIDs)
	next := 1
	for _, rosterID := range rosterIDs {
		for next == userTeamNumber(rules) {
			next++
		}
		if next <= rules.TeamCount {
			result[rosterID] = next
			next++
		}
	}
	return result
}
