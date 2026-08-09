package api

import (
	"errors"
	"net/http"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
)

type draftActionRequest struct {
	LeagueID   string       `json:"leagueId"`
	PlayerID   string       `json:"playerId"`
	Action     draft.Action `json:"action"`
	Cost       float64      `json:"cost"`
	TeamNumber int          `json:"teamNumber"`
}

type leagueDraftRequest struct {
	LeagueID string `json:"leagueId"`
}

type playerPreferenceRequest struct {
	LeagueID   string `json:"leagueId"`
	PlayerID   string `json:"playerId"`
	Preference string `json:"preference"`
}

type sleeperSyncRequest struct {
	LeagueID       string `json:"leagueId"`
	SleeperDraftID string `json:"sleeperDraftId"`
	RosterID       int    `json:"rosterId"`
}

func registerDraftRoutes(mux *http.ServeMux, drafts *application.DraftService, leagues *application.LeagueService) {
	mux.HandleFunc("GET /api/v1/draft", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.URL.Query().Get("leagueId")
		if leagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		snapshot, err := drafts.Snapshot(request.Context(), leagueID)
		if writeServiceError(response, err, http.StatusInternalServerError, "Unable to load the draft.",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("POST /api/v1/draft/actions", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[draftActionRequest](response, request, "The draft action was not valid.")
		if !ok {
			return
		}
		if input.LeagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		snapshot, err := drafts.RecordForTeam(request.Context(), input.LeagueID, input.PlayerID, input.Action, input.Cost, input.TeamNumber)
		if writeServiceError(response, err, http.StatusBadRequest, "",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."},
			serviceError{application.ErrPlayerUnavailable, http.StatusConflict, "That player is no longer available."},
			serviceError{application.ErrDraftComplete, http.StatusConflict, "This draft is complete."},
			serviceError{application.ErrDraftNotStarted, http.StatusConflict, "Start the draft before recording selections."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("POST /api/v1/draft/undo", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[leagueDraftRequest](response, request, "The undo request was not valid.")
		if !ok {
			return
		}
		if input.LeagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		snapshot, err := drafts.Undo(request.Context(), input.LeagueID)
		if writeServiceError(response, err, http.StatusInternalServerError, "Unable to undo the last action.",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."},
			serviceError{application.ErrNothingToUndo, http.StatusConflict, "There is no draft action to undo."},
			serviceError{application.ErrDraftNotStarted, http.StatusConflict, "Start the draft before undoing selections."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("POST /api/v1/draft/mock", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[leagueDraftRequest](response, request, "A league ID is required.")
		if !ok || input.LeagueID == "" {
			if ok {
				writeError(response, http.StatusBadRequest, "A league ID is required.")
			}
			return
		}
		snapshot, err := drafts.MockToNextTurn(request.Context(), input.LeagueID)
		if writeServiceError(response, err, http.StatusBadRequest, "",
			serviceError{application.ErrDraftNotStarted, http.StatusConflict, "Start the draft before simulating selections."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("POST /api/v1/draft/sync/sleeper", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[sleeperSyncRequest](response, request, "The Sleeper sync settings were not valid.")
		if !ok {
			return
		}
		result, err := drafts.SyncSleeper(request.Context(), input.LeagueID, input.SleeperDraftID, input.RosterID)
		if errors.Is(err, application.ErrDraftNotStarted) {
			writeError(response, http.StatusConflict, "Start the draft before syncing selections.")
			return
		}
		if err != nil {
			writeError(response, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, result)
	})

	mux.HandleFunc("PUT /api/v1/draft/preferences", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[playerPreferenceRequest](response, request, "The player preference was not valid.")
		if !ok {
			return
		}
		if _, err := leagues.SetPlayerPreference(request.Context(), input.LeagueID, input.PlayerID, input.Preference); writeLeagueServiceError(response, err) {
			return
		}
		snapshot, err := drafts.Snapshot(request.Context(), input.LeagueID)
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to refresh the draft board.")
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
}
