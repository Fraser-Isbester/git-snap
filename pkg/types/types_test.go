package types

import (
	"testing"
	"time"
)

func TestMatchTypeString(t *testing.T) {
	tests := []struct {
		matchType MatchType
		expected  string
	}{
		{EXACT, "exact"},
		{CONTAINS, "contains"},
		{REGEX, "regex"},
		{FUZZY, "fuzzy"},
		{MatchType(999), "unknown"},
	}

	for _, test := range tests {
		result := test.matchType.String()
		if result != test.expected {
			t.Errorf("MatchType.String() = %s, want %s", result, test.expected)
		}
	}
}

func TestSnapEventCreation(t *testing.T) {
	timestamp := time.Now()
	event := SnapEvent{
		ID:        "test-id",
		Timestamp: timestamp,
		Attributes: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	if event.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got %s", event.ID)
	}
	if !event.Timestamp.Equal(timestamp) {
		t.Errorf("Expected timestamp to match")
	}
	if event.Attributes["key1"] != "value1" {
		t.Errorf("Expected key1 to be 'value1', got %v", event.Attributes["key1"])
	}
	if event.Attributes["key2"] != 42 {
		t.Errorf("Expected key2 to be 42, got %v", event.Attributes["key2"])
	}
}

func TestEnrichedCommitCreation(t *testing.T) {
	timestamp := time.Now()
	prNumber := 123
	commit := EnrichedCommit{
		SHA:         "abc123",
		Author:      "John Doe",
		AuthorEmail: "john@example.com",
		Timestamp:   timestamp,
		Message:     "Test commit",
		Files:       []string{"file1.go", "file2.go"},
		Repository:  "test-repo",
		PRNumber:    &prNumber,
		Additions:   10,
		Deletions:   5,
		Parents:     []string{"def456"},
	}

	if commit.SHA != "abc123" {
		t.Errorf("Expected SHA to be 'abc123', got %s", commit.SHA)
	}
	if commit.Author != "John Doe" {
		t.Errorf("Expected author to be 'John Doe', got %s", commit.Author)
	}
	if *commit.PRNumber != 123 {
		t.Errorf("Expected PR number to be 123, got %d", *commit.PRNumber)
	}
	if len(commit.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(commit.Files))
	}
}
