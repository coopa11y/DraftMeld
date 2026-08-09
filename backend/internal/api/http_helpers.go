package api

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const maximumJSONBytes = 64 << 10

type serviceError struct {
	target  error
	status  int
	message string
}

func decodeJSON[T any](response http.ResponseWriter, request *http.Request, invalidMessage string) (T, bool) {
	var value T
	request.Body = http.MaxBytesReader(response, request.Body, maximumJSONBytes)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		var sizeError *http.MaxBytesError
		if errors.As(err, &sizeError) {
			writeError(response, http.StatusRequestEntityTooLarge, "The JSON request exceeds the 64 KiB limit.")
		} else {
			writeError(response, http.StatusBadRequest, invalidMessage)
		}
		return value, false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeError(response, http.StatusBadRequest, "The request must contain one JSON value.")
		return value, false
	}
	return value, true
}

func writeServiceError(response http.ResponseWriter, err error, fallbackStatus int, fallbackMessage string, known ...serviceError) bool {
	if err == nil {
		return false
	}
	for _, candidate := range known {
		if errors.Is(err, candidate.target) {
			writeError(response, candidate.status, candidate.message)
			return true
		}
	}
	if fallbackMessage == "" {
		fallbackMessage = err.Error()
	}
	writeError(response, fallbackStatus, fallbackMessage)
	return true
}

func writeError(response http.ResponseWriter, status int, message string) {
	writeJSON(response, status, map[string]string{"error": message})
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
