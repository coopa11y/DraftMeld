package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/application"
)

const maximumBackupBytes = 256 * 1024

var unsafeFilenameCharacters = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

func registerExportRoutes(mux *http.ServeMux, service *application.ExportService) {
	mux.HandleFunc("GET /api/v1/leagues/{leagueID}/backup", func(response http.ResponseWriter, request *http.Request) {
		backup, err := service.Backup(request.Context(), request.PathValue("leagueID"))
		if writeExportError(response, err) {
			return
		}
		contents, err := json.MarshalIndent(backup, "", "  ")
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to create the league backup.")
			return
		}
		contents = append(contents, '\n')
		writeDownload(response, "application/json; charset=utf-8", downloadFilename(backup.Rules.Name, "backup.json"), contents)
	})

	mux.HandleFunc("POST /api/v1/leagues/import", func(response http.ResponseWriter, request *http.Request) {
		request.Body = http.MaxBytesReader(response, request.Body, maximumBackupBytes)
		decoder := json.NewDecoder(request.Body)
		decoder.DisallowUnknownFields()
		var backup application.LeagueBackup
		if err := decoder.Decode(&backup); err != nil {
			var maximumSizeError *http.MaxBytesError
			if errors.As(err, &maximumSizeError) {
				writeError(response, http.StatusRequestEntityTooLarge, "The league backup exceeds the 256 KiB limit.")
				return
			}
			writeError(response, http.StatusBadRequest, "The league backup was not valid JSON.")
			return
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			writeError(response, http.StatusBadRequest, "The request must contain one league backup.")
			return
		}
		configuration, err := service.ImportBackup(request.Context(), backup)
		if writeExportError(response, err) {
			return
		}
		writeJSON(response, http.StatusCreated, toLeagueResponse(configuration))
	})

	mux.HandleFunc("GET /api/v1/leagues/{leagueID}/exports/rankings.csv", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.PathValue("leagueID")
		contents, err := service.RankingsCSV(request.Context(), leagueID)
		if writeExportError(response, err) {
			return
		}
		writeDownload(response, "text/csv; charset=utf-8", downloadFilename(leagueID, "rankings.csv"), contents)
	})

	mux.HandleFunc("GET /api/v1/leagues/{leagueID}/exports/draft.csv", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.PathValue("leagueID")
		contents, err := service.DraftCSV(request.Context(), leagueID)
		if writeExportError(response, err) {
			return
		}
		writeDownload(response, "text/csv; charset=utf-8", downloadFilename(leagueID, "draft.csv"), contents)
	})

	mux.HandleFunc("GET /api/v1/leagues/{leagueID}/exports/draft.json", func(response http.ResponseWriter, request *http.Request) {
		leagueID := request.PathValue("leagueID")
		exported, err := service.Draft(request.Context(), leagueID)
		if writeExportError(response, err) {
			return
		}
		contents, err := json.MarshalIndent(exported, "", "  ")
		if err != nil {
			writeError(response, http.StatusInternalServerError, "Unable to create the draft export.")
			return
		}
		contents = append(contents, '\n')
		writeDownload(response, "application/json; charset=utf-8", downloadFilename(leagueID, "draft.json"), contents)
	})
}

func writeExportError(response http.ResponseWriter, err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, application.ErrLeagueNotFound) {
		writeError(response, http.StatusNotFound, "That league was not found.")
		return true
	}
	if errors.Is(err, application.ErrInvalidLeague) || errors.Is(err, application.ErrUnsupportedBackupVersion) {
		writeError(response, http.StatusBadRequest, err.Error())
		return true
	}
	writeError(response, http.StatusInternalServerError, "Unable to export or restore league data.")
	return true
}

func writeDownload(response http.ResponseWriter, contentType, filename string, contents []byte) {
	response.Header().Set("Content-Type", contentType)
	response.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	response.Header().Set("X-Content-Type-Options", "nosniff")
	response.WriteHeader(http.StatusOK)
	_, _ = response.Write(contents)
}

func downloadFilename(name, suffix string) string {
	base := strings.Trim(unsafeFilenameCharacters.ReplaceAllString(name, "-"), "-")
	if base == "" {
		base = "draftmeld"
	}
	return strings.ToLower(base) + "-" + suffix
}
