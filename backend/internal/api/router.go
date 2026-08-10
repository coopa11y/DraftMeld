package api

import (
	"log/slog"
	"net/http"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/webui"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func NewRouter(
	logger *slog.Logger,
	version string,
	drafts *application.DraftService,
	leagues *application.LeagueService,
	rankings *application.RankingService,
	projectionServices ...*application.ProjectionService,
) http.Handler {
	mux := http.NewServeMux()
	registerHealthRoute(mux, version)
	registerDraftRoutes(mux, drafts, leagues)
	registerDraftSessionRoutes(mux, drafts)
	registerDraftTradeRoutes(mux, drafts)
	registerLeagueRoutes(mux, leagues, drafts)
	registerExportRoutes(mux, application.NewExportService(leagues, drafts, rankings))
	registerRankingRoutes(mux, rankings, leagues)
	if len(projectionServices) > 0 && projectionServices[0] != nil {
		registerProjectionRoutes(mux, projectionServices[0])
	}
	mux.Handle("/", webui.Handler())
	return requestLogger(logger, mux)
}

func registerHealthRoute(mux *http.ServeMux, version string) {
	mux.HandleFunc("GET /api/v1/health", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusOK, healthResponse{Status: "ok", Service: "draftmeld", Version: version})
	})
}
