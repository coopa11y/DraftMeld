package application

import "github.com/coopa11y/DraftMeld/backend/internal/domain/draft"

type draftState struct {
	playerActions map[string]draft.Action
	activeEvents  []draft.Event
}

func replay(events []draft.Event) draftState {
	undone := make(map[int64]bool)
	for _, event := range events {
		if event.Action == draft.ActionUndo && event.TargetEventID != nil {
			undone[*event.TargetEventID] = true
		}
	}
	state := draftState{
		playerActions: make(map[string]draft.Action),
		activeEvents:  make([]draft.Event, 0, len(events)),
	}
	for _, event := range events {
		if event.Action == draft.ActionUndo || undone[event.ID] {
			continue
		}
		state.playerActions[event.PlayerID] = event.Action
		state.activeEvents = append(state.activeEvents, event)
	}
	return state
}
