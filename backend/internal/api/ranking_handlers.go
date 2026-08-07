package api

import (
	"errors"
	"io"
	"net/http"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/document"
)

func registerRankingRoutes(mux *http.ServeMux, service *application.RankingService, leagues *application.LeagueService) {
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
	mux.HandleFunc("POST /api/v1/ranking-sources/import-pdf", func(response http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(response, request.Body, document.MaxPDFBytes+(1<<20))
		reader, err := request.MultipartReader()
		if err != nil {
			writeError(response, http.StatusBadRequest, "Upload one PDF file using the file field.")
			return
		}
		var contents []byte
		found := false
		for {
			part, nextErr := reader.NextPart()
			if nextErr == io.EOF {
				break
			}
			if nextErr != nil {
				writeError(response, http.StatusBadRequest, "The PDF upload was not valid.")
				return
			}
			if part.FormName() != "file" || part.FileName() == "" {
				_, _ = io.Copy(io.Discard, io.LimitReader(part, 1<<20))
				part.Close()
				continue
			}
			if found {
				part.Close()
				writeError(response, http.StatusBadRequest, "Upload one PDF at a time.")
				return
			}
			found = true
			contents, err = io.ReadAll(io.LimitReader(part, document.MaxPDFBytes+1))
			part.Close()
			if err != nil {
				writeError(response, http.StatusBadRequest, "The PDF could not be read.")
				return
			}
		}
		if !found {
			writeError(response, http.StatusBadRequest, "Choose a PDF file to import.")
			return
		}
		if len(contents) > document.MaxPDFBytes {
			writeError(response, http.StatusRequestEntityTooLarge, "The PDF must be 20 MiB or smaller.")
			return
		}
		result, err := service.ImportPDF(request.Context(), contents)
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusCreated, result)
	})
	mux.HandleFunc("GET /api/v1/rankings", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.URL.Query().Get("leagueId")
		if leagueID == "" {
			writeError(response, http.StatusBadRequest, "A league ID is required.")
			return
		}
		configuration, err := leagues.Get(request.Context(), leagueID)
		if err != nil {
			if errors.Is(err, application.ErrLeagueNotFound) {
				writeError(response, http.StatusNotFound, "That league was not found.")
				return
			}
			writeError(response, http.StatusInternalServerError, "Unable to load ranking weights.")
			return
		}
		rankings, err := service.Consensus(request.Context(), configuration.Rules.SourceWeights)
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to build consensus rankings.")
			return
		}
		writeJSON(response, http.StatusOK, rankings)
	})
}
