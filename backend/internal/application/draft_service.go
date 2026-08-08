package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

var (
	ErrPlayerUnavailable = errors.New("player is no longer available")
	ErrNothingToUndo     = errors.New("there is no draft action to undo")
	ErrDraftComplete     = errors.New("the draft is complete")
)

type DraftEventRepository interface {
	List(context.Context, string) ([]draft.Event, error)
	Append(context.Context, draft.Event) (draft.Event, error)
}

type seasonDraftEventRepository interface {
	ListSeason(context.Context, string, int) ([]draft.Event, error)
}

type DraftService struct {
	repository     DraftEventRepository
	players        []draft.Player
	playerByID     map[string]draft.Player
	leagues        LeagueConfigurationRepository
	rankings       *RankingService
	projections    *ProjectionService
	sleeperClient  *http.Client
	sleeperBaseURL string
	mu             sync.Mutex
}

func (service *DraftService) UseIntelligence(rankings *RankingService, projections *ProjectionService) {
	service.rankings, service.projections = rankings, projections
}

func (service *DraftService) ConfigureSleeperClient(client *http.Client, baseURL string) {
	if client != nil && baseURL != "" {
		service.sleeperClient, service.sleeperBaseURL = client, baseURL
	}
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
		sleeperClient: &http.Client{Timeout: 15 * time.Second}, sleeperBaseURL: "https://api.sleeper.app/v1",
	}, nil
}

func (service *DraftService) Snapshot(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
	if err != nil {
		return draft.Snapshot{}, fmt.Errorf("list draft events: %w", err)
	}
	events, err = resolveDraftEventAliases(ctx, service.repository, events)
	if err != nil {
		return draft.Snapshot{}, err
	}
	players, playerByID, dataMode, projectionCount, err := service.playersForLeague(ctx, configuration)
	if err != nil {
		return draft.Snapshot{}, err
	}
	trades, err := service.pickTrades(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, fmt.Errorf("list draft pick trades: %w", err)
	}
	return service.buildSnapshot(configuration, events, trades, players, playerByID, dataMode, projectionCount), nil
}

func (service *DraftService) Record(ctx context.Context, leagueID, playerID string, action draft.Action, costs ...float64) (draft.Snapshot, error) {
	cost := 0.0
	if len(costs) > 0 {
		cost = costs[0]
	}
	return service.RecordForTeam(ctx, leagueID, playerID, action, cost, 0)
}

func (service *DraftService) RecordForTeam(ctx context.Context, leagueID, playerID string, action draft.Action, cost float64, requestedTeam int) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if action != draft.ActionDraft && action != draft.ActionTaken {
		return draft.Snapshot{}, fmt.Errorf("unsupported draft action: %q", action)
	}
	if cost < 0 || (configuration.Rules.DraftType == league.DraftTypeAuction && cost <= 0) {
		return draft.Snapshot{}, errors.New("auction draft actions require a positive cost")
	}
	players, playerByID, _, _, err := service.playersForLeague(ctx, configuration)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if _, exists := playerByID[playerID]; !exists {
		return draft.Snapshot{}, fmt.Errorf("unknown player: %s", playerID)
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
	if err != nil {
		return draft.Snapshot{}, err
	}
	events, err = resolveDraftEventAliases(ctx, service.repository, events)
	if err != nil {
		return draft.Snapshot{}, err
	}
	state := replay(events)
	pickNumber := len(state.activeEvents) + 1
	if pickNumber > totalDraftPicks(configuration.Rules) {
		return draft.Snapshot{}, ErrDraftComplete
	}
	trades, err := service.pickTrades(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	teamNumber, err := teamForDraftAction(configuration.Rules, pickNumber, ownerForPick(pickNumber, configuration.Rules, trades), action, requestedTeam)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if _, unavailable := state.playerActions[playerID]; unavailable {
		return draft.Snapshot{}, ErrPlayerUnavailable
	}
	if configuration.Rules.DraftType == league.DraftTypeAuction && action == draft.ActionDraft {
		available := make([]draft.Player, 0, len(players))
		myRosterSize := 0
		for _, player := range players {
			existingAction, unavailable := state.playerActions[player.ID]
			if !unavailable {
				available = append(available, player)
			} else if existingAction == draft.ActionDraft {
				myRosterSize++
			}
		}
		auctionRules := configuration.Rules
		auctionRules.AuctionBudget += auctionBudgetAdjustments(trades, configuration.Rules.Season)[configuration.Rules.DraftPosition]
		_, _, maximumBid := auctionState(auctionRules, service.history(state.activeEvents, playerByID, configuration.Rules), available, myRosterSize)
		if cost > maximumBid {
			return draft.Snapshot{}, fmt.Errorf("bid exceeds your maximum available bid of $%.0f", maximumBid)
		}
	}
	if _, err = service.repository.Append(ctx, draft.Event{LeagueID: leagueID, Season: configuration.Rules.Season, PlayerID: playerID, Action: action, Cost: cost, TeamNumber: teamNumber}); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) history(events []draft.Event, playerByID map[string]draft.Player, rules league.Rules) []draft.Pick {
	history := make([]draft.Pick, 0, len(events))
	for index, event := range events {
		teamNumber := event.TeamNumber
		if teamNumber == 0 && rules.DraftType != league.DraftTypeAuction {
			teamNumber = pickOwner(index+1, rules)
		}
		history = append(history, draft.Pick{EventID: event.ID, Number: index + 1, Action: event.Action, Player: playerByID[event.PlayerID], CreatedAt: event.CreatedAt, Cost: event.Cost, TeamNumber: teamNumber, TeamName: teamName(rules, teamNumber)})
	}
	return history
}

func (service *DraftService) Undo(ctx context.Context, leagueID string) (draft.Snapshot, error) {
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
	state := replay(events)
	if len(state.activeEvents) == 0 {
		return draft.Snapshot{}, ErrNothingToUndo
	}
	target := state.activeEvents[len(state.activeEvents)-1]
	if _, err = service.repository.Append(ctx, draft.Event{
		LeagueID: leagueID, Season: configuration.Rules.Season, PlayerID: target.PlayerID, Action: draft.ActionUndo, TargetEventID: &target.ID,
	}); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) listDraftEvents(ctx context.Context, leagueID string, season int) ([]draft.Event, error) {
	if repository, ok := service.repository.(seasonDraftEventRepository); ok {
		return repository.ListSeason(ctx, leagueID, season)
	}
	return service.repository.List(ctx, leagueID)
}

func (service *DraftService) configuration(ctx context.Context, leagueID string) (LeagueConfiguration, error) {
	configuration, exists, err := service.leagues.GetLeague(ctx, leagueID)
	if err != nil {
		return LeagueConfiguration{}, fmt.Errorf("load league configuration: %w", err)
	}
	if !exists {
		return LeagueConfiguration{}, fmt.Errorf("%w: %s", ErrLeagueNotFound, leagueID)
	}
	configuration.Rules = withDefaultSourcePreferences(configuration.Rules)
	return configuration, nil
}

func (service *DraftService) buildSnapshot(configuration LeagueConfiguration, events []draft.Event, trades []draft.PickTrade, players []draft.Player, playerByID map[string]draft.Player, dataMode string, projectionCount int) draft.Snapshot {
	state := replay(events)
	available := make([]draft.Player, 0, len(players))
	myTeam := make([]draft.Player, 0)
	history := make([]draft.Pick, 0, len(state.activeEvents))

	for _, player := range players {
		action, unavailable := state.playerActions[player.ID]
		if !unavailable {
			available = append(available, player)
		} else if action == draft.ActionDraft {
			myTeam = append(myTeam, player)
		}
	}
	for index, event := range state.activeEvents {
		teamNumber := event.TeamNumber
		if teamNumber == 0 && configuration.Rules.DraftType != league.DraftTypeAuction {
			teamNumber = pickOwner(index+1, configuration.Rules)
		}
		history = append(history, draft.Pick{
			EventID: event.ID, Number: index + 1, Action: event.Action,
			Player: playerByID[event.PlayerID], CreatedAt: event.CreatedAt, Cost: event.Cost,
			TeamNumber: teamNumber, TeamName: teamName(configuration.Rules, teamNumber),
		})
	}

	pickNumber := len(state.activeEvents) + 1
	totalPicks := totalDraftPicks(configuration.Rules)
	complete := len(state.activeEvents) >= totalPicks
	nextPick := nextUserPickWithTrades(pickNumber-1, configuration.Rules, trades)
	auctionRules := configuration.Rules
	auctionRules.AuctionBudget += auctionBudgetAdjustments(trades, configuration.Rules.Season)[configuration.Rules.DraftPosition]
	budgetRemaining, inflation, maximumBid := auctionState(auctionRules, history, available, len(myTeam))
	teams := draftTeams(configuration.Rules, history, trades)
	onClock := 0
	if !complete && configuration.Rules.DraftType != league.DraftTypeAuction {
		onClock = ownerForPick(pickNumber, configuration.Rules, trades)
	}
	return draft.Snapshot{
		LeagueID: configuration.ID, LeagueName: configuration.Rules.Name, PickNumber: pickNumber,
		Available: available, MyTeam: myTeam, Teams: teams, History: history,
		Recommendations:     recommend(available, myTeam, configuration.Rules, configuration.Recommendation, recommendationContext{NextUserPick: nextPick, RecentPicks: history}),
		CanUndo:             len(state.activeEvents) > 0,
		DataMode:            dataMode,
		ProjectionCount:     projectionCount,
		DraftType:           string(configuration.Rules.DraftType),
		NextUserPick:        nextPick,
		AuctionBudget:       auctionRules.AuctionBudget,
		BudgetRemaining:     budgetRemaining,
		AuctionInflation:    inflation,
		AuctionMinimumBid:   configuration.Rules.AuctionMinimumBid,
		MaximumBid:          maximumBid,
		IsUserTurn:          !complete && onClock == configuration.Rules.DraftPosition,
		TotalPicks:          totalPicks,
		IsComplete:          complete,
		OnClockTeamNumber:   onClock,
		PickSlots:           draftPickSlots(configuration.Rules, trades, len(state.activeEvents)),
		PickTrades:          enrichPickTrades(configuration.Rules, trades),
		LeagueFormat:        string(configuration.Rules.LeagueFormat),
		Season:              configuration.Rules.Season,
		AuctionBudgetTrades: configuration.Rules.AuctionBudgetTrades,
	}
}

func totalDraftPicks(rules league.Rules) int {
	rosterSize := 0
	for _, slot := range rules.RosterSlots {
		rosterSize += slot.Count
	}
	return rosterSize * rules.TeamCount
}

func pickOwner(pick int, rules league.Rules) int {
	if pick < 1 || rules.TeamCount < 1 || rules.DraftType == league.DraftTypeAuction {
		return 0
	}
	round := (pick - 1) / rules.TeamCount
	slot := (pick-1)%rules.TeamCount + 1
	if rules.DraftType == league.DraftTypeSnake && round%2 == 1 {
		return rules.TeamCount - slot + 1
	}
	return slot
}

func teamName(rules league.Rules, number int) string {
	if number >= 1 && number <= len(rules.TeamNames) {
		return rules.TeamNames[number-1]
	}
	return "Unassigned"
}

func teamForDraftAction(rules league.Rules, pick, scheduledOwner int, action draft.Action, requested int) (int, error) {
	if rules.DraftType == league.DraftTypeAuction {
		if action == draft.ActionDraft {
			return rules.DraftPosition, nil
		}
		if requested < 1 || requested > rules.TeamCount || requested == rules.DraftPosition {
			return 0, errors.New("choose the opponent team that won this player")
		}
		return requested, nil
	}
	owner := scheduledOwner
	if requested != 0 {
		if requested < 1 || requested > rules.TeamCount {
			return 0, fmt.Errorf("pick owner must be between 1 and %d", rules.TeamCount)
		}
		owner = requested
	}
	if action == draft.ActionDraft && owner != rules.DraftPosition {
		return 0, fmt.Errorf("pick %d belongs to %s", pick, teamName(rules, owner))
	}
	if action == draft.ActionTaken && owner == rules.DraftPosition {
		return 0, errors.New("it is your turn; use Draft to add a player to your team")
	}
	return owner, nil
}

func draftTeams(rules league.Rules, history []draft.Pick, trades []draft.PickTrade) []draft.Team {
	teams := make([]draft.Team, rules.TeamCount)
	adjustments := auctionBudgetAdjustments(trades, rules.Season)
	for index := range teams {
		budget := 0.0
		if rules.DraftType == league.DraftTypeAuction {
			budget = rules.AuctionBudget + adjustments[index+1]
			if index+1 == rules.DraftPosition {
				budget -= rules.MyKeeperSpend
			}
		}
		teams[index] = draft.Team{Number: index + 1, Name: teamName(rules, index+1), IsUser: index+1 == rules.DraftPosition, Roster: []draft.Player{}, AuctionBudgetRemaining: budget}
	}
	for _, pick := range history {
		if pick.TeamNumber >= 1 && pick.TeamNumber <= len(teams) {
			teams[pick.TeamNumber-1].Roster = append(teams[pick.TeamNumber-1].Roster, pick.Player)
			teams[pick.TeamNumber-1].AuctionBudgetRemaining -= pick.Cost
		}
	}
	return teams
}
