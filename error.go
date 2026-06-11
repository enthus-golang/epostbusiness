package epostbusiness

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// maxErrorBodyBytes bounds how much of a non-2xx response body is read into an
// APIError, guarding against an unexpectedly large body.
const maxErrorBodyBytes = 1 << 20 // 1 MiB

// APIError is returned when the E-POSTBUSINESS API responds with a non-2xx status.
// It preserves the HTTP status code and the raw response body so callers can log
// and react to the actual provider error instead of a flattened message. When the
// body is the provider's JSON error envelope, its fields are parsed out too.
type APIError struct {
	StatusCode  int       // HTTP status code returned by the API.
	Code        string    // Provider error code (e.g. "E315"), if the body parsed.
	Level       string    // Provider error level, if the body parsed.
	Description string    // Provider error description, if the body parsed.
	Date        time.Time // Provider error timestamp, if the body parsed.
	Body        []byte    // Raw response body (bounded by maxErrorBodyBytes).
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	switch {
	case e.Code != "" && e.Description != "":
		return fmt.Sprintf("%s: %s", e.Code, e.Description)
	case e.Description != "":
		return e.Description
	case e.Code != "":
		return fmt.Sprintf("error code %s", e.Code)
	default:
		return fmt.Sprintf("epostbusiness: unexpected status %d", e.StatusCode)
	}
}

// errorEnvelope is the provider's JSON error body. Date is a string (parsed
// separately) so a missing or non-RFC3339 date does not fail the whole decode and
// drop the code/level/description.
type errorEnvelope struct {
	Level       string `json:"level"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Date        string `json:"date"`
}

// newAPIError reads the (bounded) response body and builds an APIError, parsing the
// provider's JSON error envelope when present. It always returns a non-nil
// *APIError. The caller still closes res.Body.
func newAPIError(res *http.Response) *APIError {
	if res == nil {
		return &APIError{}
	}

	var body []byte
	if res.Body != nil {
		body, _ = io.ReadAll(io.LimitReader(res.Body, maxErrorBodyBytes))
	}

	e := &APIError{StatusCode: res.StatusCode, Body: body}

	var env errorEnvelope
	if len(body) > 0 && json.Unmarshal(body, &env) == nil {
		e.Code = env.Code
		e.Level = env.Level
		e.Description = env.Description
		if env.Date != "" {
			if date, perr := time.Parse(time.RFC3339, env.Date); perr == nil {
				e.Date = date
			}
		}
	}

	return e
}
