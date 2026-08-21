package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/document"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

const maxRankingCSVBytes = 10 << 20

func registerRankingRoutes(mux *http.ServeMux, service *application.RankingService, leagues *application.LeagueService) {
	mux.HandleFunc("GET /api/v1/player-directory/status", func(response http.ResponseWriter, request *http.Request) {
		status, err := service.PlayerDirectoryStatus(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load the player directory.")
			return
		}
		writeJSON(response, http.StatusOK, status)
	})
	mux.HandleFunc("GET /api/v1/ranking-identities", func(response http.ResponseWriter, request *http.Request) {
		issues, err := service.IdentityIssues(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load the identity review queue.")
			return
		}
		writeJSON(response, http.StatusOK, issues)
	})
	mux.HandleFunc("POST /api/v1/ranking-identities/review", func(response http.ResponseWriter, request *http.Request) {
		input, ok := decodeJSON[struct {
			IssueKey           string `json:"issueKey"`
			Resolution         string `json:"resolution"`
			CanonicalPlayerKey string `json:"canonicalPlayerKey"`
		}](response, request, "The identity review was not valid.")
		if !ok {
			return
		}
		if err := service.ReviewIdentity(request.Context(), input.IssueKey, input.Resolution, input.CanonicalPlayerKey); err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		response.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /api/v1/ranking-sources", func(response http.ResponseWriter, request *http.Request) {
		sources, err := service.Sources(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load ranking sources.")
			return
		}
		writeJSON(response, http.StatusOK, sources)
	})
	mux.HandleFunc("GET /api/v1/ranking-sources/recommendations", func(response http.ResponseWriter, request *http.Request) {
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
			writeError(response, http.StatusInternalServerError, "Unable to load league rules.")
			return
		}
		recommendations, err := service.Recommendations(request.Context(), configuration.Rules)
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to recommend ranking sources.")
			return
		}
		writeJSON(response, http.StatusOK, recommendations)
	})
	mux.HandleFunc("POST /api/v1/ranking-sources/refresh", func(response http.ResponseWriter, request *http.Request) {
		var sources []ranking.SourceStatus
		var err error
		if leagueID := request.URL.Query().Get("leagueId"); leagueID != "" {
			configuration, leagueErr := leagues.Get(request.Context(), leagueID)
			if leagueErr != nil {
				if errors.Is(leagueErr, application.ErrLeagueNotFound) {
					writeError(response, http.StatusNotFound, "That league was not found.")
					return
				}
				writeError(response, http.StatusInternalServerError, "Unable to load league rules.")
				return
			}
			sources, err = service.RefreshForLeague(request.Context(), configuration.Rules)
		} else {
			sources, err = service.Refresh(request.Context())
		}
		if err != nil {
			writeError(response, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, sources)
	})
	mux.HandleFunc("POST /api/v1/ranking-sources/{sourceId}/refresh", func(response http.ResponseWriter, request *http.Request) {
		source, err := service.RefreshSource(request.Context(), request.PathValue("sourceId"))
		if err != nil {
			switch {
			case errors.Is(err, application.ErrRankingSourceNotFound):
				writeError(response, http.StatusNotFound, "That ranking source was not found.")
			case errors.Is(err, application.ErrRankingSourceNotRefreshable):
				writeError(response, http.StatusBadRequest, "That ranking source must be updated by importing a file.")
			default:
				writeError(response, http.StatusBadGateway, err.Error())
			}
			return
		}
		writeJSON(response, http.StatusOK, source)
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
	mux.HandleFunc("POST /api/v1/ranking-sources/import-csv", func(response http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(response, request.Body, maxRankingCSVBytes+(1<<20))
		if err := request.ParseMultipartForm(maxRankingCSVBytes); err != nil {
			writeError(response, http.StatusBadRequest, "Upload a ranking CSV no larger than 10 MiB.")
			return
		}
		if request.MultipartForm != nil {
			defer request.MultipartForm.RemoveAll()
		}
		name := strings.TrimSpace(request.FormValue("name"))
		file, _, err := request.FormFile("file")
		if err != nil {
			writeError(response, http.StatusBadRequest, "Choose a ranking CSV to import.")
			return
		}
		defer file.Close()
		mapping := make(map[string]string)
		if rawMapping := strings.TrimSpace(request.FormValue("mapping")); rawMapping != "" {
			if err = json.Unmarshal([]byte(rawMapping), &mapping); err != nil {
				writeError(response, http.StatusBadRequest, "The ranking column mapping was not valid.")
				return
			}
		}
		status, err := service.ImportCSV(request.Context(), name, io.LimitReader(file, maxRankingCSVBytes+1), mapping)
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusCreated, status)
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
		rankings, err := service.Consensus(request.Context(), configuration.Rules.SourcePreferences, configuration.Rules.ConsensusMethod)
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to build consensus rankings.")
			return
		}
		writeJSON(response, http.StatusOK, rankings)
	})
	mux.HandleFunc("GET /api/v1/ranking-watchlist", func(response http.ResponseWriter, request *http.Request) {
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
			writeError(response, http.StatusInternalServerError, "Unable to load ranking preferences.")
			return
		}
		players, err := service.Watchlist(request.Context(), configuration.Rules.SourcePreferences, configuration.Rules.ConsensusMethod)
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to build the disabled-source watchlist.")
			return
		}
		writeJSON(response, http.StatusOK, players)
	})
}
