package parser

import (
	"testing"
)

func TestParseJSONEvents(t *testing.T) {
	jsonData := `[
		{
			"id": "event1",
			"timestamp": "2023-01-01T12:00:00Z",
			"attributes": {
				"user_id": "john.doe",
				"action": "code_generation"
			},
			"metadata": {
				"source": "api"
			}
		},
		{
			"id": "event2",
			"timestamp": "2023-01-01T12:05:00Z",
			"attributes": {
				"user_id": "jane.smith",
				"action": "code_review"
			},
			"metadata": {
				"source": "webhook"
			}
		}
	]`

	events, err := parseJSONEvents([]byte(jsonData))
	if err != nil {
		t.Fatalf("Failed to parse JSON events: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].ID != "event1" {
		t.Errorf("Expected first event ID to be 'event1', got %s", events[0].ID)
	}

	if events[0].Attributes["user_id"] != "john.doe" {
		t.Errorf("Expected user_id to be 'john.doe', got %v", events[0].Attributes["user_id"])
	}
}

func TestParseJSONLEvents(t *testing.T) {
	jsonlData := `{"id": "event1", "timestamp": "2023-01-01T12:00:00Z", "attributes": {"user_id": "john.doe"}}
{"id": "event2", "timestamp": "2023-01-01T12:05:00Z", "attributes": {"user_id": "jane.smith"}}
`

	events, err := parseJSONLEvents([]byte(jsonlData))
	if err != nil {
		t.Fatalf("Failed to parse JSONL events: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].ID != "event1" {
		t.Errorf("Expected first event ID to be 'event1', got %s", events[0].ID)
	}

	if events[1].ID != "event2" {
		t.Errorf("Expected second event ID to be 'event2', got %s", events[1].ID)
	}
}

func TestParseCSVEvents(t *testing.T) {
	csvData := `id,timestamp,user_id,action
event1,2023-01-01T12:00:00Z,john.doe,code_generation
event2,2023-01-01T12:05:00Z,jane.smith,code_review`

	events, err := parseCSVEvents([]byte(csvData))
	if err != nil {
		t.Fatalf("Failed to parse CSV events: %v", err)
	}

	if len(events) != 2 {
		t.Errorf("Expected 2 events, got %d", len(events))
	}

	if events[0].ID != "event1" {
		t.Errorf("Expected first event ID to be 'event1', got %s", events[0].ID)
	}

	if events[0].Attributes["user_id"] != "john.doe" {
		t.Errorf("Expected user_id to be 'john.doe', got %v", events[0].Attributes["user_id"])
	}

	if events[0].Attributes["action"] != "code_generation" {
		t.Errorf("Expected action to be 'code_generation', got %v", events[0].Attributes["action"])
	}
}

func TestParseTimestamp(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"2023-01-01T12:00:00Z", true},
		{"2023-01-01T12:00:00.123Z", true},
		{"2023-01-01 12:00:00", true},
		{"2023-01-01T12:00:00", true},
		{"invalid-timestamp", false},
		{"", false},
	}

	for _, tc := range testCases {
		_, err := parseTimestamp(tc.input)
		if tc.expected && err != nil {
			t.Errorf("Expected timestamp %s to be valid, got error: %v", tc.input, err)
		}
		if !tc.expected && err == nil {
			t.Errorf("Expected timestamp %s to be invalid, but it was parsed successfully", tc.input)
		}
	}
}

func TestParseValue(t *testing.T) {
	testCases := []struct {
		input    string
		expected interface{}
	}{
		{"42", 42},
		{"3.14", 3.14},
		{"true", true},
		{"false", false},
		{"hello", "hello"},
		{"", ""},
		{" trimmed ", "trimmed"},
	}

	for _, tc := range testCases {
		result := parseValue(tc.input)
		if result != tc.expected {
			t.Errorf("parseValue(%s) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestDetermineFormat(t *testing.T) {
	testCases := []struct {
		filename string
		expected string
	}{
		{"events.json", "json"},
		{"events.jsonl", "jsonl"},
		{"events.csv", "csv"},
		{"events.JSON", "json"},
		{"events.JSONL", "jsonl"},
		{"events.CSV", "csv"},
		{"events.txt", "json"},
		{"events", "json"},
	}

	for _, tc := range testCases {
		result := determineFormat(tc.filename)
		if result != tc.expected {
			t.Errorf("determineFormat(%s) = %s, want %s", tc.filename, result, tc.expected)
		}
	}
}
