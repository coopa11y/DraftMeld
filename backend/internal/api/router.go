package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
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
	LeagueID   string       `json:"leagueId"`
	PlayerID   string       `json:"playerId"`
	Action     draft.Action `json:"action"`
	Cost       float64      `json:"cost"`
	TeamNumber int          `json:"teamNumber"`
}

type undoRequest struct {
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

type pickTradeRequest struct {
	LeagueID        string `json:"leagueId"`
	TeamOneNumber   int    `json:"teamOneNumber"`
	TeamTwoNumber   int    `json:"teamTwoNumber"`
	TeamOneReceives []int  `json:"teamOneReceives"`
	TeamTwoReceives []int  `json:"teamTwoReceives"`
}

func NewRouter(
	logger *slog.Logger,
	version string,
	draftService *application.DraftService,
	leagueService *application.LeagueService,
	rankingService *application.RankingService,
	projectionServices ...*application.ProjectionService,
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
		snapshot, err := draftService.RecordForTeam(request.Context(), input.LeagueID, input.PlayerID, input.Action, input.Cost, input.TeamNumber)
		if errors.Is(err, application.ErrLeagueNotFound) {
			writeError(response, http.StatusNotFound, "That league was not found.")
			return
		}
		if errors.Is(err, application.ErrPlayerUnavailable) {
			writeError(response, http.StatusConflict, "That player is no longer available.")
			return
		}
		if errors.Is(err, application.ErrDraftComplete) {
			writeError(response, http.StatusConflict, "This draft is complete.")
			return
		}
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
	mux.HandleFunc("POST /api/v1/draft/pick-trades", func(response http.ResponseWriter, request *http.Request) {
		var input pickTradeRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil || input.LeagueID == "" {
			writeError(response, http.StatusBadRequest, "The draft-pick trade was not valid.")
			return
		}
		snapshot, err := draftService.CreatePickTrade(request.Context(), input.LeagueID, input.TeamOneNumber, input.TeamTwoNumber, input.TeamOneReceives, input.TeamTwoReceives)
		if errors.Is(err, application.ErrLeagueNotFound) {
			writeError(response, http.StatusNotFound, "That league was not found.")
			return
		}
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusCreated, snapshot)
	})
	mux.HandleFunc("DELETE /api/v1/draft/pick-trades/{tradeId}", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.URL.Query().Get("leagueId")
		tradeID, err := strconv.ParseInt(request.PathValue("tradeId"), 10, 64)
		if leagueID == "" || err != nil || tradeID < 1 {
			writeError(response, http.StatusBadRequest, "A league ID and trade ID are required.")
			return
		}
		snapshot, err := draftService.DeletePickTrade(request.Context(), leagueID, tradeID)
		if errors.Is(err, application.ErrLeagueNotFound) {
			writeError(response, http.StatusNotFound, "That league was not found.")
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
	mux.HandleFunc("POST /api/v1/draft/mock", func(response http.ResponseWriter, request *http.Request) {
		var input undoRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil || input.LeagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		snapshot, err := draftService.MockToNextTurn(request.Context(), input.LeagueID)
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
	mux.HandleFunc("POST /api/v1/draft/sync/sleeper", func(response http.ResponseWriter, request *http.Request) {
		var input sleeperSyncRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			writeError(response, http.StatusBadRequest, "The Sleeper sync settings were not valid.")
			return
		}
		result, err := draftService.SyncSleeper(request.Context(), input.LeagueID, input.SleeperDraftID, input.RosterID)
		if err != nil {
			writeError(response, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, result)
	})
	mux.HandleFunc("PUT /api/v1/draft/preferences", func(response http.ResponseWriter, request *http.Request) {
		var input playerPreferenceRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			writeError(response, http.StatusBadRequest, "The player preference was not valid.")
			return
		}
		if _, err := leagueService.SetPlayerPreference(request.Context(), input.LeagueID, input.PlayerID, input.Preference); err != nil {
			if writeLeagueServiceError(response, err) {
				return
			}
		}
		snapshot, err := draftService.Snapshot(request.Context(), input.LeagueID)
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to refresh the draft board.")
			return
		}
		writeJSON(response, http.StatusOK, snapshot)
	})
	registerLeagueRoutes(mux, leagueService)
	registerExportRoutes(mux, application.NewExportService(leagueService, draftService, rankingService))
	registerRankingRoutes(mux, rankingService, leagueService)
	if len(projectionServices) > 0 && projectionServices[0] != nil {
		registerProjectionRoutes(mux, projectionServices[0])
	}
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
