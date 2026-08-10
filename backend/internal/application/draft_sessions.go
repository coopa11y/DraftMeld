package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

var (
	ErrDraftNotStarted         = errors.New("the draft has not started")
	ErrDraftStarted            = errors.New("the draft has already started")
	ErrDraftPositionUnassigned = errors.New("set your draft position in league setup before starting the draft")
	ErrNoResetToUndo           = errors.New("there is no draft reset to undo")
)

type DraftSessionRepository interface {
	GetDraftSession(context.Context, string, int) (draft.Session, bool, error)
	SaveDraftSession(context.Context, draft.Session) error
	ResetDraftSession(context.Context, draft.Session) error
	RestoreDraftSession(context.Context, draft.Session, []draft.Event) error
}

func (service *DraftService) StartDraft(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if configuration.Rules.DraftType != league.DraftTypeAuction && configuration.Rules.DraftPosition == 0 {
		return draft.Snapshot{}, ErrDraftPositionUnassigned
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
	if err != nil {
		return draft.Snapshot{}, err
	}
	state := replay(events)
	session, err := service.draftSession(ctx, leagueID, configuration.Rules.Season, len(state.activeEvents), totalDraftPicks(configuration.Rules))
	if err != nil {
		return draft.Snapshot{}, err
	}
	if session.Status != draft.SessionNotStarted {
		return draft.Snapshot{}, ErrDraftStarted
	}
	repository, ok := service.repository.(DraftSessionRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft session storage is unavailable")
	}
	now := time.Now().UTC()
	session = draft.Session{
		LeagueID: leagueID, Season: configuration.Rules.Season, Status: draft.SessionInProgress,
		StartedAt: now, UpdatedAt: now, ResetEvents: []draft.Event{},
	}
	if err = repository.SaveDraftSession(ctx, session); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) ResetDraft(ctx context.Context, leagueID, confirmation string) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if confirmation != configuration.Rules.Name {
		return draft.Snapshot{}, errors.New("type the league name exactly to confirm the reset")
	}
	events, err := service.listDraftEvents(ctx, leagueID, configuration.Rules.Season)
	if err != nil {
		return draft.Snapshot{}, err
	}
	state := replay(events)
	session, err := service.draftSession(ctx, leagueID, configuration.Rules.Season, len(state.activeEvents), totalDraftPicks(configuration.Rules))
	if err != nil {
		return draft.Snapshot{}, err
	}
	if session.Status == draft.SessionNotStarted {
		return draft.Snapshot{}, ErrDraftNotStarted
	}
	repository, ok := service.repository.(DraftSessionRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft session storage is unavailable")
	}
	backup := make([]draft.Event, len(state.activeEvents))
	for index, event := range state.activeEvents {
		backup[index] = draft.Event{
			LeagueID: leagueID, Season: configuration.Rules.Season, PlayerID: event.PlayerID,
			Action: event.Action, Cost: event.Cost, TeamNumber: event.TeamNumber,
		}
	}
	previousStatus := session.Status
	session.Status = draft.SessionNotStarted
	session.ResetStatus = previousStatus
	session.ResetEvents = backup
	session.StartedAt = time.Time{}
	if err = repository.ResetDraftSession(ctx, session); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) UndoDraftReset(ctx context.Context, leagueID string) (draft.Snapshot, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	configuration, err := service.configuration(ctx, leagueID)
	if err != nil {
		return draft.Snapshot{}, err
	}
	repository, ok := service.repository.(DraftSessionRepository)
	if !ok {
		return draft.Snapshot{}, errors.New("draft session storage is unavailable")
	}
	session, found, err := repository.GetDraftSession(ctx, leagueID, configuration.Rules.Season)
	if err != nil {
		return draft.Snapshot{}, err
	}
	if !found || session.Status != draft.SessionNotStarted || session.ResetStatus == "" {
		return draft.Snapshot{}, ErrNoResetToUndo
	}
	backup := append([]draft.Event(nil), session.ResetEvents...)
	session.Status = session.ResetStatus
	session.ResetStatus = ""
	session.ResetEvents = []draft.Event{}
	if session.Status != draft.SessionNotStarted {
		session.StartedAt = time.Now().UTC()
	}
	if err = repository.RestoreDraftSession(ctx, session, backup); err != nil {
		return draft.Snapshot{}, err
	}
	return service.Snapshot(ctx, leagueID)
}

func (service *DraftService) draftSession(ctx context.Context, leagueID string, season, activeEvents, totalPicks int) (draft.Session, error) {
	repository, ok := service.repository.(DraftSessionRepository)
	if !ok {
		return draft.Session{LeagueID: leagueID, Season: season, Status: draftStatus(activeEvents, totalPicks, draft.SessionInProgress)}, nil
	}
	session, found, err := repository.GetDraftSession(ctx, leagueID, season)
	if err != nil {
		return draft.Session{}, fmt.Errorf("load draft session: %w", err)
	}
	if !found {
		session = draft.Session{LeagueID: leagueID, Season: season, Status: draft.SessionNotStarted, ResetEvents: []draft.Event{}}
	}
	session.Status = draftStatus(activeEvents, totalPicks, session.Status)
	return session, nil
}

func draftStatus(activeEvents, totalPicks int, stored draft.SessionStatus) draft.SessionStatus {
	if totalPicks > 0 && activeEvents >= totalPicks {
		return draft.SessionComplete
	}
	if activeEvents > 0 {
		return draft.SessionInProgress
	}
	if stored == draft.SessionInProgress {
		return draft.SessionInProgress
	}
	return draft.SessionNotStarted
}

func (service *DraftService) requireDraftStarted(ctx context.Context, leagueID string, season, activeEvents, totalPicks int) error {
	session, err := service.draftSession(ctx, leagueID, season, activeEvents, totalPicks)
	if err != nil {
		return err
	}
	if session.Status == draft.SessionNotStarted {
		return ErrDraftNotStarted
	}
	if session.Status == draft.SessionComplete {
		return ErrDraftComplete
	}
	return nil
}
