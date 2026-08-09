package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONAcceptsOneStrictValue(t *testing.T) {
	type requestBody struct {
		LeagueID string `json:"leagueId"`
	}
	tests := []struct {
		name   string
		body   string
		status int
		ok     bool
	}{
		{name: "valid", body: `{"leagueId":"demo"}`, status: http.StatusNoContent, ok: true},
		{name: "unknown field", body: `{"leagueId":"demo","extra":true}`, status: http.StatusBadRequest},
		{name: "multiple values", body: `{"leagueId":"demo"} {"leagueId":"other"}`, status: http.StatusBadRequest},
		{name: "malformed", body: `{`, status: http.StatusBadRequest},
		{name: "too large", body: `{"leagueId":"` + strings.Repeat("x", maximumJSONBytes) + `"}`, status: http.StatusRequestEntityTooLarge},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			decoded, ok := decodeJSON[requestBody](response, request, "invalid")
			if ok != test.ok {
				t.Fatalf("unexpected decode result: ok=%v body=%q", ok, response.Body.String())
			}
			if ok {
				response.WriteHeader(http.StatusNoContent)
				if decoded.LeagueID != "demo" {
					t.Fatalf("unexpected decoded value: %#v", decoded)
				}
			}
			if response.Code != test.status {
				t.Fatalf("expected status %d, got %d: %s", test.status, response.Code, response.Body.String())
			}
		})
	}
}

func TestWriteServiceErrorMapsKnownAndFallbackErrors(t *testing.T) {
	target := errors.New("known")
	response := httptest.NewRecorder()
	if !writeServiceError(response, fmtWrapped(target), http.StatusInternalServerError, "fallback",
		serviceError{target: target, status: http.StatusConflict, message: "known message"}) || response.Code != http.StatusConflict {
		t.Fatalf("known error was not mapped: %d %s", response.Code, response.Body.String())
	}

	response = httptest.NewRecorder()
	if !writeServiceError(response, errors.New("unexpected"), http.StatusBadGateway, "fallback") || response.Code != http.StatusBadGateway {
		t.Fatalf("fallback error was not mapped: %d %s", response.Code, response.Body.String())
	}
	if writeServiceError(httptest.NewRecorder(), nil, http.StatusInternalServerError, "fallback") {
		t.Fatal("nil error should not write a response")
	}
}

func fmtWrapped(target error) error {
	return errors.Join(errors.New("context"), target)
}
