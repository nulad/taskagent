package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
