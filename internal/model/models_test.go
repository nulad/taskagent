package model

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// ValidStatus should accept all five known statuses and reject anything else.
// Cover: each known value (true), empty string, unknown value, wrong case ("Todo"),
// underscore form ("in_progress") to lock in that hyphen is the contract.
func TestValidStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "todo",
			input: "todo",
			want:  true,
		},
		{
			name:  "backlog",
			input: "backlog",
			want:  true,
		},
		{
			name:  "in-progress",
			input: "in-progress",
			want:  true,
		},
		{
			name:  "review",
			input: "review",
			want:  true,
		},
		{
			name:  "done",
			input: "done",
			want:  true,
		},
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "unknown value",
			input: "unknown",
			want:  false,
		},
		{
			name:  "wrong case",
			input: "Todo",
			want:  false,
		},
		{
			name:  "underscore form",
			input: "in_progress",
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidStatus(tt.input)
			if result != tt.want {
				t.Fatalf("ValidStatus(%q) = %v, want %v", tt.input, result, tt.want)
			}
		})
	}
}

// Drift test: every TaskStatus constant must be recognized by ValidStatus.
// If someone adds a new constant later but forgets to update ValidStatus,
// this test fails. Loop over all known constants, assert ValidStatus returns true.
func TestValidStatus_AllConstantsRecognized(t *testing.T) {
	all := []TaskStatus{
		StatusTodo,
		StatusBacklog,
		StatusInProgress,
		StatusReview,
		StatusDone,
	}
	for _, s := range all {
		if !ValidStatus(s.String()) {
			t.Fatalf("ValidStatus(%q) = false, want true", s.String())
		}
	}
}

// String() should return the underlying string value.
// Quick sanity check; one assertion is enough.
func TestTaskStatus_String(t *testing.T) {
	if StatusTodo.String() != "todo" {
		t.Fatalf("StatusTodo.String() = %q, want %q", StatusTodo.String(), "todo")
	}
}

// Round-trip: marshal to JSON and unmarshal back, verify the whole struct
// matches via reflect.DeepEqual. Catches tag typos, type mismatches.
func TestProject_JSONRoundTrip(t *testing.T) {
	project := Project{
		ID:          "p-123",
		Name:        "Test Project",
		Description: "A project for testing",
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(project)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Project
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(project, unmarshaled) {
		t.Fatalf("Project round-trip failed: got %v, want %v", unmarshaled, project)
	}
}

func TestTask_JSONRoundTrip(t *testing.T) {
	task := Task{
		ID:          "t-123",
		ProjectID:   "p-123",
		Title:       "Test Task",
		Description: "A task for testing",
		Status:      StatusInProgress,
		Tags:        []string{"tag1", "tag2"},
		CreatedAt:   "2024-01-01T00:00:00Z",
		UpdatedAt:   "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Task
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(task, unmarshaled) {
		t.Fatalf("Task round-trip failed: got %v, want %v", unmarshaled, task)
	}
}

func TestApiKey_JSONRoundTrip(t *testing.T) {
	apiKey := ApiKey{
		ID:        1,
		Label:     "Test API Key",
		UserID:    42,
		UserName:  "testuser",
		CreatedAt: "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(apiKey)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ApiKey
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(apiKey, unmarshaled) {
		t.Fatalf("ApiKey round-trip failed: got %v, want %v", unmarshaled, apiKey)
	}
}

func TestUser_JSONRoundTrip(t *testing.T) {
	user := User{
		ID:        42,
		Name:      "testuser",
		Active:    true,
		IsAdmin:   false,
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled User
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(user, unmarshaled) {
		t.Fatalf("User round-trip failed: got %v, want %v", unmarshaled, user)
	}
}

// Status JSON contract: the wire format must use hyphenated values
// (matches the seed data in migrations/001_initial.sql and the React client).
// Marshal a Task with StatusInProgress, assert "in-progress" appears in output.
func TestTaskStatus_JSONFormat(t *testing.T) {
	statusInProgress := StatusInProgress

	data, err := json.Marshal(statusInProgress)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if !strings.Contains(string(data), "in-progress") {
		t.Fatalf("Expected 'in-progress' in JSON output, got %s", string(data))
	}
}

// Omitempty contract for Task: minimal task should not include optional fields
// in JSON output, but should include required ones.
func TestTask_OmitsOptionalFields(t *testing.T) {
	task := Task{
		ID:        "t-1",
		Title:     "x",
		Status:    StatusBacklog,
		CreatedAt: "2024-01-01T00:00:00Z",
		UpdatedAt: "2024-01-02T00:00:00Z",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	jsonStr := string(data)

	if strings.Contains(jsonStr, `"description"`) {
		t.Fatal("Expected 'description' to be omitted from JSON output")
	}
	if strings.Contains(jsonStr, `"tags"`) {
		t.Fatal("Expected 'tags' to be omitted from JSON output")
	}
	if !strings.Contains(jsonStr, `"id":"t-1"`) {
		t.Fatal("Expected 'id' to be present in JSON output")
	}
	if !strings.Contains(jsonStr, `"title":"x"`) {
		t.Fatal("Expected 'title' to be present in JSON output")
	}
	if !strings.Contains(jsonStr, `"status":"backlog"`) {
		t.Fatal("Expected 'status' to be present in JSON output")
	}
	if !strings.Contains(jsonStr, `"created_at":"2024-01-01T00:00:00Z"`) {
		t.Fatal("Expected 'created_at' to be present in JSON output")
	}
	if !strings.Contains(jsonStr, `"updated_at":"2024-01-02T00:00:00Z"`) {
		t.Fatal("Expected 'updated_at' to be present in JSON output")
	}
}
