package application

import (
	"context"
	"sort"
	"sync"
)

type MemoryLeagueRepository struct {
	mu      sync.RWMutex
	leagues map[string]LeagueConfiguration
}

func NewMemoryLeagueRepository(configurations ...LeagueConfiguration) *MemoryLeagueRepository {
	repository := &MemoryLeagueRepository{leagues: make(map[string]LeagueConfiguration, len(configurations))}
	for _, configuration := range configurations {
		repository.leagues[configuration.ID] = configuration
	}
	return repository
}

func (repository *MemoryLeagueRepository) ListLeagues(context.Context) ([]LeagueConfiguration, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	leagues := make([]LeagueConfiguration, 0, len(repository.leagues))
	for _, configuration := range repository.leagues {
		leagues = append(leagues, configuration)
	}
	sort.Slice(leagues, func(left, right int) bool {
		return leagues[left].Rules.Name < leagues[right].Rules.Name
	})
	return leagues, nil
}

func (repository *MemoryLeagueRepository) GetLeague(_ context.Context, id string) (LeagueConfiguration, bool, error) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()
	configuration, found := repository.leagues[id]
	return configuration, found, nil
}

func (repository *MemoryLeagueRepository) SaveLeague(_ context.Context, configuration LeagueConfiguration) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	repository.leagues[configuration.ID] = configuration
	return nil
}

func (repository *MemoryLeagueRepository) DeleteLeague(_ context.Context, id string) (bool, error) {
	repository.mu.Lock()
	defer repository.mu.Unlock()
	if _, found := repository.leagues[id]; !found {
		return false, nil
	}
	delete(repository.leagues, id)
	return true, nil
}
