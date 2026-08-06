package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/draft"
	"github.com/coopa11y/DraftMeld/backend/internal/webui"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

type actionRequest struct {
	LeagueID string       `json:"leagueId"`
	PlayerID string       `json:"playerId"`
	Action   draft.Action `json:"action"`
}

type undoRequest struct {
	LeagueID string `json:"leagueId"`
}

func NewRouter(
	logger *slog.Logger,
	version string,
	draftService *application.DraftService,
	leagueService *application.LeagueService,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusOK, healthResponse{
			Status:  "ok",
			Service: "draftmeld",
			Version: version,
		})
	})
	mux.HandleFunc("GET /api/v1/draft", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.URL.Query().Get("leagueId")
		if leagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		snapshot, err := draftService.Snapshot(request.Context(), leagueID)
		if errors.Is(err, application.ErrLeagueNotFound) {
			writeError(response, http.StatusNotFound, "That league was not found.")
			return
		}
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load the draft.")
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
	mux.HandleFunc("POST /api/v1/draft/actions", func(response http.ResponseWriter, request *http.Request) {
		var input actionRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			writeError(response, http.StatusBadRequest, "The draft action was not valid.")
			return
		}
		if input.LeagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		snapshot, err := draftService.Record(request.Context(), input.LeagueID, input.PlayerID, input.Action)
		if errors.Is(err, application.ErrLeagueNotFound) {
			writeError(response, http.StatusNotFound, "That league was not found.")
			return
		}
		if errors.Is(err, application.ErrPlayerUnavailable) {
			writeError(response, http.StatusConflict, "That player is no longer available.")
			return
		}
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
	mux.HandleFunc("POST /api/v1/draft/undo", func(response http.ResponseWriter, request *http.Request) {
		var input undoRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			writeError(response, http.StatusBadRequest, "The undo request was not valid.")
			return
		}
		if input.LeagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		snapshot, err := draftService.Undo(request.Context(), input.LeagueID)
		if errors.Is(err, application.ErrLeagueNotFound) {
			writeError(response, http.StatusNotFound, "That league was not found.")
			return
		}
		if errors.Is(err, application.ErrNothingToUndo) {
			writeError(response, http.StatusConflict, "There is no draft action to undo.")
			return
		}
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to undo the last action.")
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
	registerLeagueRoutes(mux, leagueService)
	mux.Handle("/", webui.Handler())
	return requestLogger(logger, mux)
}

func writeError(response http.ResponseWriter, status int, message string) {
	writeJSON(response, status, map[string]string{"error": message})
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(response, request)
		logger.Info("request", "method", request.Method, "path", request.URL.Path, "duration", time.Since(started))
	})
}
