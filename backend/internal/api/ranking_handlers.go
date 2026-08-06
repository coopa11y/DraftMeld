package api

import (
	"net/http"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
)

func registerRankingRoutes(mux *http.ServeMux, service *application.RankingService) {
	mux.HandleFunc("GET /api/v1/ranking-sources", func(response http.ResponseWriter, request *http.Request) {
		sources, err := service.Sources(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load ranking sources.")
			return
		}
		writeJSON(response, http.StatusOK, sources)
	})
	mux.HandleFunc("POST /api/v1/ranking-sources/refresh", func(response http.ResponseWriter, request *http.Request) {
		sources, err := service.Refresh(request.Context())
		if err != nil {
			writeError(response, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, sources)
	})
	mux.HandleFunc("GET /api/v1/rankings", func(response http.ResponseWriter, request *http.Request) {
		rankings, err := service.Consensus(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to build consensus rankings.")
			return
		}
		writeJSON(response, http.StatusOK, rankings)
	})
}
