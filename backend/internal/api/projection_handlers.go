package api

import (
	"io"
	"net/http"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
)

const maxProjectionCSVBytes = 10 << 20

func registerProjectionRoutes(mux *http.ServeMux, service *application.ProjectionService) {
	mux.HandleFunc("GET /api/v1/projection-sources", func(response http.ResponseWriter, request *http.Request) {
		sources, err := service.Sources(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load projection sources.")
			return
		}
		writeJSON(response, http.StatusOK, sources)
	})
	mux.HandleFunc("POST /api/v1/projection-sources/import-csv", func(response http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(response, request.Body, maxProjectionCSVBytes+(1<<20))
		if err := request.ParseMultipartForm(maxProjectionCSVBytes); err != nil {
			writeError(response, http.StatusBadRequest, "Upload a projection CSV no larger than 10 MiB.")
			return
		}
		name := strings.TrimSpace(request.FormValue("name"))
		file, _, err := request.FormFile("file")
		if err != nil {
			writeError(response, http.StatusBadRequest, "Choose a projection CSV to import.")
			return
		}
		defer file.Close()
		status, err := service.ImportCSV(request.Context(), name, io.LimitReader(file, maxProjectionCSVBytes+1))
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusCreated, status)
	})
}
