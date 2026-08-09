package api

import (
	"net/http"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type resetDraftRequest struct {
	LeagueID     string `json:"leagueId"`
	Confirmation string `json:"confirmation"`
}

type seasonRolloverRequest struct {
	LeagueID   string           `json:"leagueId"`
	Season     int              `json:"season"`
	DraftType  league.DraftType `json:"draftType"`
	DraftOrder []int            `json:"draftOrder"`
}

func registerDraftSessionRoutes(mux *http.ServeMux, drafts *application.DraftService) {
	mux.HandleFunc("POST /api/v1/draft/session/start", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[leagueDraftRequest](response, request, "A league ID is required.")
		if !ok || input.LeagueID == "" {
			if ok {
				writeError(response, http.StatusBadRequest, "A league ID is required.")
			}
			return
		}
		snapshot, err := drafts.StartDraft(request.Context(), input.LeagueID)
		if writeServiceError(response, err, http.StatusInternalServerError, "Unable to start the draft.",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."},
			serviceError{application.ErrDraftStarted, http.StatusConflict, "This draft has already started."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("POST /api/v1/draft/session/reset", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[resetDraftRequest](response, request, "The reset request was not valid.")
		if !ok || input.LeagueID == "" {
			if ok {
				writeError(response, http.StatusBadRequest, "The reset request was not valid.")
			}
			return
		}
		snapshot, err := drafts.ResetDraft(request.Context(), input.LeagueID, input.Confirmation)
		if writeServiceError(response, err, http.StatusBadRequest, "",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."},
			serviceError{application.ErrDraftNotStarted, http.StatusConflict, "There is no active draft to reset."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("POST /api/v1/draft/session/undo-reset", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[leagueDraftRequest](response, request, "A league ID is required.")
		if !ok || input.LeagueID == "" {
			if ok {
				writeError(response, http.StatusBadRequest, "A league ID is required.")
			}
			return
		}
		snapshot, err := drafts.UndoDraftReset(request.Context(), input.LeagueID)
		if writeServiceError(response, err, http.StatusInternalServerError, "Unable to restore the draft.",
			serviceError{application.ErrLeagueNotFound, http.StatusNotFound, "That league was not found."},
			serviceError{application.ErrNoResetToUndo, http.StatusConflict, "There is no draft reset to undo."}) {
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})

	mux.HandleFunc("POST /api/v1/draft/seasons", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[seasonRolloverRequest](response, request, "The next-season setup was not valid.")
		if !ok || input.LeagueID == "" {
			if ok {
				writeError(response, http.StatusBadRequest, "The next-season setup was not valid.")
			}
			return
		}
		snapshot, err := drafts.AdvanceSeason(request.Context(), input.LeagueID, input.Season, input.DraftType, input.DraftOrder)
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
}
