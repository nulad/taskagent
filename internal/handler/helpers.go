package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
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
