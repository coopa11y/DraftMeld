package api

import (
	"errors"
	"net/http"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
)

type addNewsSourceRequest struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	Attribution    string `json:"attribution"`
	RefreshMinutes int    `json:"refreshMinutes"`
}

type updateNewsSourceRequest struct {
	Enabled bool `json:"enabled"`
}

func registerPlayerNewsRoutes(mux *http.ServeMux, service *application.PlayerNewsService) {
	mux.HandleFunc("GET /api/v1/player-news", func(response http.ResponseWriter, request *http.Request) {
		feed, err := service.Feed(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load player news.")
			return
		}
		writeJSON(response, http.StatusOK, feed)
	})
	mux.HandleFunc("POST /api/v1/player-news/refresh", func(response http.ResponseWriter, request *http.Request) {
		result := service.Refresh(request.Context(), true)
		status := http.StatusOK
		if result.Refreshed == 0 && len(result.Errors) > 0 {
			status = http.StatusBadGateway
		}
		writeJSON(response, status, result)
	})
	mux.HandleFunc("GET /api/v1/player-news/sources", func(response http.ResponseWriter, request *http.Request) {
		sources, err := service.Sources(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load player news sources.")
			return
		}
		writeJSON(response, http.StatusOK, sources)
	})
	mux.HandleFunc("POST /api/v1/player-news/sources", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[addNewsSourceRequest](response, request, "The RSS source was not valid.")
		if !ok {
			return
		}
		source, err := service.AddRSSSource(request.Context(), input.Name, input.URL, input.Attribution, input.RefreshMinutes)
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusCreated, source)
	})
	mux.HandleFunc("PATCH /api/v1/player-news/sources/{sourceId}", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[updateNewsSourceRequest](response, request, "The source setting was not valid.")
		if !ok {
			return
		}
		if err := service.SetSourceEnabled(request.Context(), request.PathValue("sourceId"), input.Enabled); err != nil {
			if errors.Is(err, application.ErrNewsSourceNotFound) {
				writeError(response, http.StatusNotFound, "That player news source was not found.")
				return
			}
			writeError(response, http.StatusInternalServerError, "Unable to update the player news source.")
			return
		}
		response.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("DELETE /api/v1/player-news/sources/{sourceId}", func(response http.ResponseWriter, request *http.Request) {
		if err := service.DeleteSource(request.Context(), request.PathValue("sourceId")); err != nil {
			switch {
			case errors.Is(err, application.ErrBuiltInNewsSource):
				writeError(response, http.StatusBadRequest, "Built-in sources can be disabled but not deleted.")
			case errors.Is(err, application.ErrNewsSourceNotFound):
				writeError(response, http.StatusNotFound, "That player news source was not found.")
			default:
				writeError(response, http.StatusInternalServerError, "Unable to delete the player news source.")
			}
			return
		}
		response.WriteHeader(http.StatusNoContent)
	})
}
