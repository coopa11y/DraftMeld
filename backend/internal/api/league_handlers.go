package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
	"github.com/coopa11y/DraftMeld/backend/internal/domain/league"
)

type leagueResponse struct {
	ID string `json:"id"`
	league.Rules
}

func registerLeagueRoutes(mux *http.ServeMux, service *application.LeagueService) {
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

	mux.HandleFunc("DELETE /api/v1/leagues/{leagueID}", func(response http.ResponseWriter, request *http.Request) {
		err := service.Delete(request.Context(), request.PathValue("leagueID"))
		if writeLeagueServiceError(response, err) {
			return
		}
		response.WriteHeader(http.StatusNoContent)
	})
}

func decodeLeagueRules(response http.ResponseWriter, request *http.Request) (league.Rules, bool) {
	request.Body = http.MaxBytesReader(response, request.Body, 64*1024)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	var rules league.Rules
	if err := decoder.Decode(&rules); err != nil {
		writeError(response, http.StatusBadRequest, "The league settings were not valid JSON.")
		return league.Rules{}, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(response, http.StatusBadRequest, "The request must contain one league configuration.")
		return league.Rules{}, false
	}
	return rules, true
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
