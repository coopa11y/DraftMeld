package application

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

var (
	ErrInvalidLeague  = errors.New("league configuration is invalid")
	ErrLeagueNotFound = errors.New("league configuration was not found")
)

type LeagueConfigurationRepository interface {
	ListLeagues(context.Context) ([]LeagueConfiguration, error)
	GetLeague(context.Context, string) (LeagueConfiguration, bool, error)
	SaveLeague(context.Context, LeagueConfiguration) error
	DeleteLeague(context.Context, string) (bool, error)
}

type LeagueService struct {
	repository LeagueConfigurationRepository
	mu         sync.Mutex
}

func NewLeagueService(repository LeagueConfigurationRepository) *LeagueService {
	return &LeagueService{repository: repository}
}

func (service *LeagueService) EnsureDefault(ctx context.Context) error {
	leagues, err := service.repository.ListLeagues(ctx)
	if err != nil {
		return fmt.Errorf("list leagues: %w", err)
	}
	if len(leagues) > 0 {
		return nil
	}
	if err = service.repository.SaveLeague(ctx, DemoLeagueConfiguration()); err != nil {
		return fmt.Errorf("create default league: %w", err)
	}
	return nil
}

func (service *LeagueService) List(ctx context.Context) ([]LeagueConfiguration, error) {
	configurations, err := service.repository.ListLeagues(ctx)
	if err != nil {
		return nil, err
	}
	for index := range configurations {
		configurations[index].Rules = withDefaultSourcePreferences(configurations[index].Rules)
	}
	return configurations, nil
}

func (service *LeagueService) Get(ctx context.Context, id string) (LeagueConfiguration, error) {
	configuration, found, err := service.repository.GetLeague(ctx, id)
	if err != nil {
		return LeagueConfiguration{}, err
	}
	if !found {
		return LeagueConfiguration{}, fmt.Errorf("%w: %s", ErrLeagueNotFound, id)
	}
	configuration.Rules = withDefaultSourcePreferences(configuration.Rules)
	return configuration, nil
}

func (service *LeagueService) Create(ctx context.Context, rules league.Rules) (LeagueConfiguration, error) {
	return service.create(ctx, rules, DefaultRecommendationPolicy())
}

func (service *LeagueService) Restore(ctx context.Context, rules league.Rules, recommendation league.RecommendationPolicy) (LeagueConfiguration, error) {
	return service.create(ctx, rules, recommendation)
}

func (service *LeagueService) create(ctx context.Context, rules league.Rules, recommendation league.RecommendationPolicy) (LeagueConfiguration, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	rules = withDefaultSourcePreferences(rules)
	configuration := LeagueConfiguration{
		ID: slugify(rules.Name), Rules: rules, Recommendation: recommendation,
	}
	if err := configuration.Validate(); err != nil {
		return LeagueConfiguration{}, fmt.Errorf("%w: %v", ErrInvalidLeague, err)
	}
	id, err := service.availableID(ctx, configuration.ID)
	if err != nil {
		return LeagueConfiguration{}, err
	}
	configuration.ID = id
	if err = service.repository.SaveLeague(ctx, configuration); err != nil {
		return LeagueConfiguration{}, err
	}
	return configuration, nil
}

func (service *LeagueService) Update(ctx context.Context, id string, rules league.Rules) (LeagueConfiguration, error) {
	service.mu.Lock()
	defer service.mu.Unlock()

	configuration, err := service.Get(ctx, id)
	if err != nil {
		return LeagueConfiguration{}, err
	}
	if configuration.Rules.LeagueFormat == league.LeagueFormatDynasty && rules.Season != configuration.Rules.Season {
		return LeagueConfiguration{}, fmt.Errorf("%w: use the guided season rollover to change a dynasty league's season", ErrInvalidLeague)
	}
	updatedRules := withDefaultSourcePreferences(rules)
	locked, lockErr := service.draftStructureLocked(ctx, id, configuration.Rules.Season)
	if lockErr != nil {
		return LeagueConfiguration{}, lockErr
	}
	if locked && draftStructureChanged(configuration.Rules, updatedRules) {
		return LeagueConfiguration{}, fmt.Errorf("%w: reset the current draft before changing its structure", ErrInvalidLeague)
	}
	configuration.Rules = updatedRules
	if err = configuration.Validate(); err != nil {
		return LeagueConfiguration{}, fmt.Errorf("%w: %v", ErrInvalidLeague, err)
	}
	if err = service.repository.SaveLeague(ctx, configuration); err != nil {
		return LeagueConfiguration{}, err
	}
	return configuration, nil
}

func (service *LeagueService) draftStructureLocked(ctx context.Context, leagueID string, season int) (bool, error) {
	if sessions, ok := service.repository.(DraftSessionRepository); ok {
		session, found, err := sessions.GetDraftSession(ctx, leagueID, season)
		if err != nil {
			return false, err
		}
		if found && (session.Status == draft.SessionInProgress || session.Status == draft.SessionComplete) {
			return true, nil
		}
	}
	if events, ok := service.repository.(seasonDraftEventRepository); ok {
		seasonEvents, err := events.ListSeason(ctx, leagueID, season)
		if err != nil {
			return false, err
		}
		return len(replay(seasonEvents).activeEvents) > 0, nil
	}
	return false, nil
}

func draftStructureChanged(current, updated league.Rules) bool {
	updated.Name = current.Name
	updated.TeamNames = current.TeamNames
	updated.SourcePreferences = current.SourcePreferences
	updated.ConsensusMethod = current.ConsensusMethod
	updated.PlayerPreferences = current.PlayerPreferences
	return !reflect.DeepEqual(current, updated)
}

func (service *LeagueService) Duplicate(ctx context.Context, id string) (LeagueConfiguration, error) {
	source, err := service.Get(ctx, id)
	if err != nil {
		return LeagueConfiguration{}, err
	}
	rules := source.Rules
	rules.Name = "Copy of " + rules.Name
	rules.RosterSlots = cloneRosterSlots(rules.RosterSlots)
	rules.ScoringRules = cloneScoringRules(rules.ScoringRules)
	rules.SourcePreferences = cloneSourcePreferences(rules.SourcePreferences)
	rules.PlayerPreferences = clonePlayerPreferences(rules.PlayerPreferences)
	if rules.ConsensusMethod == "" {
		rules.ConsensusMethod = "weighted-median"
	}
	if rules.PlayerPreferences == nil {
		rules.PlayerPreferences = make(map[string]string)
	}
	if rules.AuctionBudget <= 0 {
		rules.AuctionBudget = 200
	}
	if rules.AuctionMinimumBid <= 0 {
		rules.AuctionMinimumBid = 1
	}
	return service.Create(ctx, rules)
}

func (service *LeagueService) Delete(ctx context.Context, id string) error {
	deleted, err := service.repository.DeleteLeague(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return fmt.Errorf("%w: %s", ErrLeagueNotFound, id)
	}
	return nil
}

func (service *LeagueService) SetPlayerPreference(ctx context.Context, id, playerID, preference string) (LeagueConfiguration, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.Get(ctx, id)
	if err != nil {
		return LeagueConfiguration{}, err
	}
	if playerID == "" || (preference != "" && preference != "target" && preference != "avoid") {
		return LeagueConfiguration{}, fmt.Errorf("%w: player preference must be target, avoid, or empty", ErrInvalidLeague)
	}
	configuration.Rules.PlayerPreferences = clonePlayerPreferences(configuration.Rules.PlayerPreferences)
	if preference == "" {
		delete(configuration.Rules.PlayerPreferences, playerID)
	} else {
		configuration.Rules.PlayerPreferences[playerID] = preference
	}
	if err = service.repository.SaveLeague(ctx, configuration); err != nil {
		return LeagueConfiguration{}, err
	}
	return configuration, nil
}

func (service *LeagueService) availableID(ctx context.Context, base string) (string, error) {
	for suffix := 1; suffix < 10_000; suffix++ {
		candidate := base
		if suffix > 1 {
			candidate = fmt.Sprintf("%s-%d", base, suffix)
		}
		_, found, err := service.repository.GetLeague(ctx, candidate)
		if err != nil {
			return "", err
		}
		if !found {
			return candidate, nil
		}
	}
	return "", errors.New("unable to allocate a league ID")
}

var nonSlugCharacter = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(name string) string {
	slug := strings.Trim(nonSlugCharacter.ReplaceAllString(strings.ToLower(name), "-"), "-")
	if slug == "" {
		return "league"
	}
	return slug
}

func cloneRosterSlots(slots []league.RosterSlot) []league.RosterSlot {
	cloned := make([]league.RosterSlot, len(slots))
	for index, slot := range slots {
		cloned[index] = slot
		cloned[index].Positions = append([]string(nil), slot.Positions...)
	}
	return cloned
}

func cloneScoringRules(rules map[string]float64) map[string]float64 {
	cloned := make(map[string]float64, len(rules))
	for name, value := range rules {
		cloned[name] = value
	}
	return cloned
}

func cloneSourcePreferences(preferences map[string]league.RankingSourcePreference) map[string]league.RankingSourcePreference {
	cloned := make(map[string]league.RankingSourcePreference, len(preferences))
	for sourceID, preference := range preferences {
		cloned[sourceID] = preference
	}
	return cloned
}

func clonePlayerPreferences(preferences map[string]string) map[string]string {
	cloned := make(map[string]string, len(preferences))
	for playerID, preference := range preferences {
		cloned[playerID] = preference
	}
	return cloned
}

func withDefaultSourcePreferences(rules league.Rules) league.Rules {
	if rules.UserTeamNumber == 0 {
		rules.UserTeamNumber = rules.DraftPosition
	}
	if len(rules.DraftOrder) != rules.TeamCount {
		rules.DraftOrder = make([]int, rules.TeamCount)
		for index := range rules.DraftOrder {
			rules.DraftOrder[index] = index + 1
		}
	}
	if position := draftPositionForTeam(rules.DraftOrder, rules.UserTeamNumber); position > 0 {
		rules.DraftPosition = position
	}
	rules.TeamNames = normalizedTeamNames(rules.TeamNames, rules.TeamCount, rules.UserTeamNumber)
	defaults := DefaultRankingSourcePreferences()
	rules.SourcePreferences = cloneSourcePreferences(rules.SourcePreferences)
	for sourceID, preference := range defaults {
		if _, exists := rules.SourcePreferences[sourceID]; !exists {
			rules.SourcePreferences[sourceID] = preference
		}
	}
	if rules.ConsensusMethod == "" {
		rules.ConsensusMethod = "weighted-median"
	}
	if rules.PlayerPreferences == nil {
		rules.PlayerPreferences = make(map[string]string)
	}
	if rules.AuctionBudget <= 0 {
		rules.AuctionBudget = 200
	}
	if rules.AuctionMinimumBid <= 0 {
		rules.AuctionMinimumBid = 1
	}
	if rules.LeagueFormat == "" {
		rules.LeagueFormat = league.LeagueFormatRedraft
	}
	if rules.Season == 0 {
		rules.Season = time.Now().UTC().Year()
	}
	if rules.InitialSeason == 0 {
		rules.InitialSeason = rules.Season
	}
	if rules.LeagueFormat == league.LeagueFormatRedraft {
		rules.FuturePickSeasons = 0
	} else if rules.FuturePickSeasons == 0 {
		rules.FuturePickSeasons = 3
	}
	if rules.RookieDraftRounds == 0 {
		rules.RookieDraftRounds = 4
	}
	if rules.LeagueFormat == league.LeagueFormatDynasty && rules.FAABBudget == 0 {
		rules.FAABBudget = 100
	}
	return rules
}

func draftPositionForTeam(order []int, teamNumber int) int {
	for index, team := range order {
		if team == teamNumber {
			return index + 1
		}
	}
	return 0
}

func normalizedTeamNames(names []string, teamCount, userPosition int) []string {
	result := make([]string, teamCount)
	for index := range result {
		if index < len(names) {
			result[index] = strings.TrimSpace(names[index])
		}
		if result[index] == "" {
			result[index] = fmt.Sprintf("Team %d", index+1)
		}
	}
	if userPosition >= 1 && userPosition <= teamCount && (len(names) < userPosition || strings.TrimSpace(names[userPosition-1]) == "") {
		result[userPosition-1] = "My Team"
	}
	return result
}
