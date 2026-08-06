package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/coopa11y/DraftMeld/backend/internal/webui"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

func NewRouter(logger *slog.Logger, version string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(response http.ResponseWriter, _ *http.Request) {
		writeJSON(response, http.StatusOK, healthResponse{
			Status:  "ok",
			Service: "draftmeld",
			Version: version,
		})
	})
	mux.Handle("/", webui.Handler())
	return requestLogger(logger, mux)
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
