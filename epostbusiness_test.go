package epostbusiness

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// newTestAPI returns an API whose HTTP client answers every request with the given
// status and body, so the error paths can be exercised without network access.
func newTestAPI(status int, body string) *API {
	api := New()
	api.Client = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})}
	return api
}

func TestAPIErrorError(t *testing.T) {
	cases := []struct {
		name string
		err  *APIError
		want string
	}{
		{"with code and description", &APIError{Code: "E315", Description: "Fehler in Sendungsverarbeitung"}, "E315: Fehler in Sendungsverarbeitung"},
		{"description only", &APIError{Description: "Fehler in Sendungsverarbeitung"}, "Fehler in Sendungsverarbeitung"},
		{"code only", &APIError{Code: "E315"}, "error code E315"},
		{"empty envelope falls back to status", &APIError{StatusCode: http.StatusBadGateway}, "epostbusiness: unexpected status 502"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.err.Error(); got != c.want {
				t.Fatalf("Error() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestNewAPIErrorParsesEnvelope(t *testing.T) {
	const body = `{"level":"Error","code":"E315","description":"Fehler in Sendungsverarbeitung","date":"2026-06-11T08:45:58Z"}`
	e := newAPIError(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(body))})

	if e.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", e.StatusCode)
	}
	if e.Code != "E315" {
		t.Errorf("Code = %q, want E315", e.Code)
	}
	if e.Level != "Error" {
		t.Errorf("Level = %q, want Error", e.Level)
	}
	if e.Description != "Fehler in Sendungsverarbeitung" {
		t.Errorf("Description = %q", e.Description)
	}
	if e.Date.IsZero() {
		t.Error("Date was not parsed")
	}
	if string(e.Body) != body {
		t.Errorf("Body = %q, want the raw body", e.Body)
	}
}

func TestNewAPIErrorMalformedDate(t *testing.T) {
	// A missing or non-RFC3339 date must not drop the rest of the envelope.
	const body = `{"level":"Error","code":"E315","description":"Fehler","date":"not-a-date"}`
	e := newAPIError(&http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader(body))})

	if e.Code != "E315" {
		t.Errorf("Code = %q, want E315 (must survive a bad date)", e.Code)
	}
	if e.Description != "Fehler" {
		t.Errorf("Description = %q, want Fehler", e.Description)
	}
	if !e.Date.IsZero() {
		t.Errorf("Date = %v, want zero for an unparseable date", e.Date)
	}
}

func TestNewAPIErrorNilResponse(t *testing.T) {
	if e := newAPIError(nil); e == nil {
		t.Fatal("newAPIError(nil) returned nil, want non-nil *APIError")
	}

	e := newAPIError(&http.Response{StatusCode: http.StatusBadGateway})
	if e.StatusCode != http.StatusBadGateway {
		t.Errorf("StatusCode = %d, want 502", e.StatusCode)
	}
	if len(e.Body) != 0 {
		t.Errorf("Body = %q, want empty for a nil response body", e.Body)
	}
}

func TestNewAPIErrorNonJSONBody(t *testing.T) {
	const body = "<html>502 Bad Gateway</html>"
	e := newAPIError(&http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader(body))})

	if e.StatusCode != http.StatusBadGateway {
		t.Errorf("StatusCode = %d, want 502", e.StatusCode)
	}
	if e.Code != "" || e.Description != "" {
		t.Errorf("expected empty envelope fields, got code=%q description=%q", e.Code, e.Description)
	}
	if string(e.Body) != body {
		t.Errorf("Body = %q, want the raw body", e.Body)
	}
}

func TestNewAPIErrorBoundsBody(t *testing.T) {
	big := strings.Repeat("a", maxErrorBodyBytes+1024)
	e := newAPIError(&http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader(big))})

	if len(e.Body) != maxErrorBodyBytes {
		t.Errorf("Body length = %d, want %d (bounded)", len(e.Body), maxErrorBodyBytes)
	}
}

func TestLoginReturnsAPIError(t *testing.T) {
	const body = `{"level":"Error","code":"E001","description":"invalid credentials"}`
	api := newTestAPI(http.StatusUnauthorized, body)

	ok, err := api.Login(context.Background(), "vendor", "ekp", "secret", "password")
	if ok {
		t.Error("ok = true, want false")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("StatusCode = %d, want 401", apiErr.StatusCode)
	}
	if apiErr.Code != "E001" {
		t.Errorf("Code = %q, want E001", apiErr.Code)
	}
	if string(apiErr.Body) != body {
		t.Error("raw Body was not preserved")
	}
}

func TestLoginSuccess(t *testing.T) {
	api := newTestAPI(http.StatusOK, `{"token":"jwt-123"}`)

	ok, err := api.Login(context.Background(), "vendor", "ekp", "secret", "password")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Error("ok = false, want true")
	}
	if api.jwt != "jwt-123" {
		t.Errorf("jwt = %q, want jwt-123", api.jwt)
	}
}

func TestCreateLettersReturnsAPIError(t *testing.T) {
	const body = `{"level":"Error","code":"E315","description":"Fehler in Sendungsverarbeitung"}`
	api := newTestAPI(http.StatusBadRequest, body)

	_, err := api.CreateLetters(context.Background(), []Letter{{}})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %v", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.Code != "E315" {
		t.Errorf("Code = %q, want E315", apiErr.Code)
	}
}

func TestGetLettersStatusListReturnsAPIError(t *testing.T) {
	api := newTestAPI(http.StatusTooManyRequests, `{"code":"E429","description":"rate limited"}`)

	_, err := api.GetLettersStatusList(context.Background(), []int{1})
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error is not *APIError: %v", err)
	}
	if apiErr.StatusCode != http.StatusTooManyRequests {
		t.Errorf("StatusCode = %d, want 429", apiErr.StatusCode)
	}
}
