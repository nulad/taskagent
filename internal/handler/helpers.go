package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
)

var (
	ErrMultipleJSONObjects = errors.New("body must contain a single JSON object")
	ErrEncodeJSONFailed    = errors.New("failed to encode JSON response")
)

func readJSON(r *http.Request, dst any) error {
	const maxJSONBodyBytes = 1 * 1024 * 1024 // 1 MiB

	body := http.MaxBytesReader(nil, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(dst)
	if err != nil {
		return err
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return ErrMultipleJSONObjects
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "failed to encode JSON response"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func pathParam(r *http.Request, key string) string {
	return r.PathValue(key)
}

// validationErrors collects field-specific validation errors.
// The zero value is ready to use; the underlying map is created on first add.
type validationErrors struct {
	errs map[string]string
}

// newValidationErrors creates a new validationErrors instance ready for use.
func newValidationErrors() *validationErrors {
	return &validationErrors{errs: make(map[string]string)}
}

// add registers a field-level error message.
func (v *validationErrors) add(field, message string) {
	v.errs[field] = message
}

// hasErrors reports whether any field errors have been collected.
func (v *validationErrors) hasErrors() bool {
	return len(v.errs) > 0
}

// Error returns a single stable message string describing all collected
// field errors, sorted by field name for deterministic output.
// Satisfies the error interface.
func (v *validationErrors) Error() string {
	if len(v.errs) == 0 {
		return "validation failed"
	}
	fields := make([]string, 0, len(v.errs))
	for field := range v.errs {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	var parts []string
	for _, field := range fields {
		parts = append(parts, field+" "+v.errs[field])
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// writeValidationErrors writes a 400 Bad Request response containing the
// human-readable validation message as a single JSON {"error":"..."} object.
func writeValidationErrors(w http.ResponseWriter, verrs *validationErrors) {
	writeError(w, http.StatusBadRequest, verrs.Error())
}
