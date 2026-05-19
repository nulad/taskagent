package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidationErrors(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		verrs := newValidationErrors()

		if verrs.hasErrors() {
			t.Errorf("expected no errors, but hasErrors() returned true")
		}

		wantErr := "validation failed"
		if verrs.Error() != wantErr {
			t.Errorf("Error() = %q, want %q", verrs.Error(), wantErr)
		}
	})

	t.Run("single field error", func(t *testing.T) {
		verrs := newValidationErrors()
		verrs.add("email", "is required")

		if !verrs.hasErrors() {
			t.Errorf("expected errors, but hasErrors() returned false")
		}

		want := "validation failed: email is required"
		if verrs.Error() != want {
			t.Errorf("Error() = %q, want %q", verrs.Error(), want)
		}
	})

	t.Run("multiple field errors", func(t *testing.T) {
		verrs := newValidationErrors()
		verrs.add("password", "is too short")
		verrs.add("email", "is invalid")
		verrs.add("name", "is required")

		if !verrs.hasErrors() {
			t.Errorf("expected errors, but hasErrors() returned false")
		}

		// Output should be sorted by field name for determinism
		want := "validation failed: email is invalid; name is required; password is too short"
		if verrs.Error() != want {
			t.Errorf("Error() = %q, want %q", verrs.Error(), want)
		}
	})

	t.Run("satisfies error interface", func(t *testing.T) {
		var _ error = newValidationErrors()
	})
}

func TestWriteValidationErrors(t *testing.T) {
	tests := []struct {
		name     string
		addFunc  func(*validationErrors)
		wantCode int
		wantBody string
	}{
		{
			name: "empty errors",
			addFunc: func(v *validationErrors) {
				// no errors added
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"validation failed"}`,
		},
		{
			name: "single field error",
			addFunc: func(v *validationErrors) {
				v.add("email", "is required")
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"validation failed: email is required"}`,
		},
		{
			name: "multiple field errors",
			addFunc: func(v *validationErrors) {
				v.add("password", "too short")
				v.add("email", "invalid")
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"validation failed: email invalid; password too short"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verrs := newValidationErrors()
			tt.addFunc(verrs)

			w := httptest.NewRecorder()
			writeValidationErrors(w, verrs)

			if w.Code != tt.wantCode {
				t.Errorf("status code = %d, want %d", w.Code, tt.wantCode)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}

			// Verify response is valid JSON with exactly one "error" key
			var resp map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("response is not valid JSON: %v\nbody: %s", err, w.Body.String())
			}
			if _, ok := resp["error"]; !ok {
				t.Errorf("response missing \"error\" key: %s", w.Body.String())
			}
			if len(resp) != 1 {
				t.Errorf("response has %d keys, want 1: %s", len(resp), w.Body.String())
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

type wantErrKind int

const (
	wantNoError wantErrKind = iota
	wantErrMultipleObjects
	wantErrMaxBytes
	wantErrUnknownField
	wantErrSyntax
)

func TestReadJSON(t *testing.T) {
	big := strings.Repeat("a", 1<<20) // 1 MiB of data

	type TestStruct struct {
		KeyOne string `json:"key_one"`
		KeyTwo string `json:"key_two"`
	}
	tests := []struct {
		name        string
		input       string
		wantErrKind wantErrKind
	}{
		{
			name:        "valid JSON object",
			input:       `{"key_one": "value_one", "key_two": "value_two"}`,
			wantErrKind: wantNoError,
		},
		{
			name:        "multiple JSON objects",
			input:       `{"key_one": "value_one", "key_two": "value_two"} {"another": "object"}`,
			wantErrKind: wantErrMultipleObjects,
		},
		{
			name:        "invalid JSON",
			input:       `{"key_one": "value_one", "key_two": invalid}`,
			wantErrKind: wantErrSyntax,
		},
		{
			name:        "oversized body",
			input:       `{"key_one":"` + big + `", "key_two": "value_two"}`,
			wantErrKind: wantErrMaxBytes,
		},
		{
			name:        "unknown field",
			input:       `{"key_one": "value_one", "key_three": "value_three"}`,
			wantErrKind: wantErrUnknownField,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.input))
			dst := &TestStruct{}
			err := readJSON(req, dst)

			switch tt.wantErrKind {
			case wantNoError:
				if err != nil {
					t.Fatalf("expected success but got error: %v", err)
				}
				expected := &TestStruct{
					KeyOne: "value_one",
					KeyTwo: "value_two",
				}
				if *dst != *expected {
					t.Errorf("expected %+v but got %+v", expected, dst)
				}

			case wantErrMultipleObjects:
				if !errors.Is(err, ErrMultipleJSONObjects) {
					t.Fatalf("expected ErrMultipleJSONObjects, got %T: %v", err, err)
				}

			case wantErrMaxBytes:
				var maxErr *http.MaxBytesError
				if !errors.As(err, &maxErr) {
					t.Fatalf("expected MaxBytesError, got %T: %v", err, err)
				}

			case wantErrSyntax:
				var syntaxErr *json.SyntaxError
				if !errors.As(err, &syntaxErr) {
					t.Fatalf("expected SyntaxError, got %T: %v", err, err)
				}

			case wantErrUnknownField:
				if !strings.Contains(err.Error(), "unknown field") {
					t.Fatalf("expected unknown field error, got %v", err)
				}
			}
		})
	}
}
