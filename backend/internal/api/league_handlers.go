package api

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/document"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type leagueResponse struct {
	ID string `json:"id"`
	league.Rules
}

type mockDraftResponse struct {
	ID       string `json:"id"`
	LeagueID string `json:"leagueId"`
	Name     string `json:"name"`
}

func registerLeagueRoutes(mux *http.ServeMux, service *application.LeagueService, drafts *application.DraftService) {
	registerLeagueRuleImportRoute(mux, application.NewLeagueRuleImportService())
	mux.HandleFunc("GET /api/v1/leagues", func(response http.ResponseWriter, request *http.Request) {
		configurations, err := service.List(request.Context())
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to load leagues.")
			return
		}
		leagues := make([]leagueResponse, 0, len(configurations))
		for _, configuration := range configurations {
			leagues = append(leagues, toLeagueResponse(configuration))
		}
		writeJSON(response, http.StatusOK, leagues)
	})

	mux.HandleFunc("POST /api/v1/leagues", func(response http.ResponseWriter, request *http.Request) {
		rules, ok := decodeLeagueRules(response, request)
		if !ok {
			return
		}
		configuration, err := service.Create(request.Context(), rules)
		if writeLeagueServiceError(response, err) {
			return
		}
		writeJSON(response, http.StatusCreated, toLeagueResponse(configuration))
	})

	mux.HandleFunc("GET /api/v1/leagues/{leagueID}", func(response http.ResponseWriter, request *http.Request) {
		configuration, err := service.Get(request.Context(), request.PathValue("leagueID"))
		if writeLeagueServiceError(response, err) {
			return
		}
		writeJSON(response, http.StatusOK, toLeagueResponse(configuration))
	})

	mux.HandleFunc("PUT /api/v1/leagues/{leagueID}", func(response http.ResponseWriter, request *http.Request) {
		rules, ok := decodeLeagueRules(response, request)
		if !ok {
			return
		}
		configuration, err := service.Update(request.Context(), request.PathValue("leagueID"), rules)
		if writeLeagueServiceError(response, err) {
			return
		}
		writeJSON(response, http.StatusOK, toLeagueResponse(configuration))
	})

	mux.HandleFunc("POST /api/v1/leagues/{leagueID}/duplicate", func(response http.ResponseWriter, request *http.Request) {
		configuration, err := service.Duplicate(request.Context(), request.PathValue("leagueID"))
		if writeLeagueServiceError(response, err) {
			return
		}
		writeJSON(response, http.StatusCreated, toLeagueResponse(configuration))
	})

	mux.HandleFunc("POST /api/v1/leagues/{leagueID}/mock-drafts", func(response http.ResponseWriter, request *http.Request) {
		configuration, err := service.CreateMockDraft(request.Context(), request.PathValue("leagueID"))
		if writeLeagueServiceError(response, err) {
			return
		}
		if _, err = drafts.StartDraft(request.Context(), configuration.ID); err != nil {
			_ = service.Delete(request.Context(), configuration.ID)
			writeError(response, http.StatusInternalServerError, "Unable to start the mock draft.")
			return
		}
		writeJSON(response, http.StatusCreated, mockDraftResponse{
			ID: configuration.ID, LeagueID: configuration.ParentLeagueID, Name: configuration.Rules.Name,
		})
	})

	mux.HandleFunc("DELETE /api/v1/leagues/{leagueID}", func(response http.ResponseWriter, request *http.Request) {
		err := service.Delete(request.Context(), request.PathValue("leagueID"))
		if writeLeagueServiceError(response, err) {
			return
		}
		response.WriteHeader(http.StatusNoContent)
	})
}

func registerLeagueRuleImportRoute(mux *http.ServeMux, service *application.LeagueRuleImportService) {
	mux.HandleFunc("POST /api/v1/leagues/rules/import", func(response http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(response, request.Body, document.MaxPDFBytes+(1<<20))
		if err := request.ParseMultipartForm(document.MaxPDFBytes); err != nil {
			writeError(response, http.StatusBadRequest, "Upload one PDF or CSV file no larger than 20 MiB.")
			return
		}
		file, header, err := request.FormFile("file")
		if err != nil {
			writeError(response, http.StatusBadRequest, "Choose a league rules PDF or CSV to import.")
			return
		}
		defer file.Close()
		contents, err := io.ReadAll(io.LimitReader(file, document.MaxPDFBytes+1))
		if err != nil || len(contents) > document.MaxPDFBytes {
			writeError(response, http.StatusRequestEntityTooLarge, "The league rules file must be 20 MiB or smaller.")
			return
		}
		extension := strings.ToLower(filepath.Ext(header.Filename))
		var result application.LeagueRuleImportResult
		switch extension {
		case ".pdf":
			result, err = service.ImportPDF(contents)
		case ".csv":
			result, err = service.ImportCSV(strings.NewReader(string(contents)))
		default:
			writeError(response, http.StatusBadRequest, "League rules must be uploaded as a PDF or CSV file.")
			return
		}
		if errors.Is(err, application.ErrNoLeagueRulesFound) {
			writeError(response, http.StatusUnprocessableEntity, "No supported league settings or scoring rules were recognized. Try a Setting/Value or Statistic/Points CSV, or configure the league manually.")
			return
		}
		if err != nil {
			writeError(response, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(response, http.StatusOK, result)
	})
}

func decodeLeagueRules(response http.ResponseWriter, request *http.Request) (league.Rules, bool) {
	return decodeJSON[league.Rules](response, request, "The league settings were not valid JSON.")
}

func writeLeagueServiceError(response http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, application.ErrLeagueNotFound) {
		writeError(response, http.StatusNotFound, "That league was not found.")
		return true
	}
	if errors.Is(err, application.ErrInvalidLeague) {
		writeError(response, http.StatusBadRequest, err.Error())
		return true
	}
	writeError(response, http.StatusInternalServerError, "Unable to save the league settings.")
	return true
}

func toLeagueResponse(configuration application.LeagueConfiguration) leagueResponse {
	return leagueResponse{ID: configuration.ID, Rules: configuration.Rules}
}
