package correlation

import (
	"testing"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/types"
)

func TestCorrelationEngine_SnapToCommits(t *testing.T) {
	config := types.SnapConfig{
		TimeWindow: 30 * time.Minute,
		AttributeRules: []types.AttributeRule{
			{
				EventKey:  "user_id",
				CommitKey: "author_email",
				MatchType: types.EXACT,
				Required:  true,
			},
		},
		ScoreWeights: map[string]float64{
			"temporal":  0.5,
			"attribute": 0.5,
		},
	}

	engine := NewCorrelationEngine(config)

	baseTime := time.Now()
	events := []types.SnapEvent{
		{
			ID:        "event1",
			Timestamp: baseTime,
			Attributes: map[string]interface{}{
				"user_id": "john.doe@example.com",
				"action":  "code_generation",
			},
		},
	}

	commits := []types.EnrichedCommit{
		{
			SHA:         "abc123",
			Author:      "John Doe",
			AuthorEmail: "john.doe@example.com",
			Timestamp:   baseTime.Add(10 * time.Minute),
			Message:     "Add new feature",
			Repository:  "test-repo",
		},
		{
			SHA:         "def456",
			Author:      "Jane Smith",
			AuthorEmail: "jane.smith@example.com",
			Timestamp:   baseTime.Add(5 * time.Minute),
			Message:     "Fix bug",
			Repository:  "test-repo",
		},
	}

	results := engine.SnapToCommits(events, commits)

	if len(results) != 1 {
		t.Errorf("Expected 1 correlation result, got %d", len(results))
	}

	if results[0].Commit.SHA != "abc123" {
		t.Errorf("Expected correlated commit SHA to be 'abc123', got %s", results[0].Commit.SHA)
	}

	if results[0].Score <= 0 {
		t.Errorf("Expected positive correlation score, got %f", results[0].Score)
	}
}

func TestCorrelationEngine_CalculateTimeDelta(t *testing.T) {
	engine := &CorrelationEngine{}

	baseTime := time.Now()
	event := types.SnapEvent{
		Timestamp: baseTime,
	}

	commit := types.EnrichedCommit{
		Timestamp: baseTime.Add(10 * time.Minute),
	}

	delta := engine.calculateTimeDelta(event, commit)
	expected := 10 * time.Minute

	if delta != expected {
		t.Errorf("Expected time delta to be %v, got %v", expected, delta)
	}

	commit2 := types.EnrichedCommit{
		Timestamp: baseTime.Add(-5 * time.Minute),
	}

	delta2 := engine.calculateTimeDelta(event, commit2)
	expected2 := 5 * time.Minute

	if delta2 != expected2 {
		t.Errorf("Expected time delta to be %v, got %v", expected2, delta2)
	}
}

func TestCorrelationEngine_MatchValues(t *testing.T) {
	engine := &CorrelationEngine{}

	testCases := []struct {
		eventValue  string
		commitValue string
		matchType   types.MatchType
		expected    bool
	}{
		{"john.doe", "john.doe", types.EXACT, true},
		{"john.doe", "jane.smith", types.EXACT, false},
		{"john", "john.doe@example.com", types.CONTAINS, true},
		{"jane", "john.doe@example.com", types.CONTAINS, false},
		{"test.*", "test-repo", types.REGEX, true},
		{"prod.*", "test-repo", types.REGEX, false},
		{"john", "john", types.FUZZY, true},
		{"xyz", "john.doe", types.FUZZY, false},
	}

	for _, tc := range testCases {
		result := engine.matchValues(tc.eventValue, tc.commitValue, tc.matchType)
		if result != tc.expected {
			t.Errorf("matchValues(%s, %s, %v) = %v, want %v",
				tc.eventValue, tc.commitValue, tc.matchType, result, tc.expected)
		}
	}
}

func TestCorrelationEngine_LevenshteinDistance(t *testing.T) {
	engine := &CorrelationEngine{}

	testCases := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "axc", 1},
		{"kitten", "sitting", 3},
	}

	for _, tc := range testCases {
		result := engine.levenshteinDistance(tc.s1, tc.s2)
		if result != tc.expected {
			t.Errorf("levenshteinDistance(%s, %s) = %d, want %d", tc.s1, tc.s2, result, tc.expected)
		}
	}
}

func TestCorrelationEngine_GetCommitValue(t *testing.T) {
	engine := &CorrelationEngine{}

	prNumber := 123
	commit := types.EnrichedCommit{
		SHA:         "abc123",
		Author:      "John Doe",
		AuthorEmail: "john.doe@example.com",
		Message:     "Test commit",
		Repository:  "test-repo",
		Branch:      "main",
		PRNumber:    &prNumber,
		Additions:   10,
		Deletions:   5,
		Files:       []string{"file1.go", "file2.go"},
	}

	testCases := []struct {
		key      string
		expected string
	}{
		{"sha", "abc123"},
		{"author", "John Doe"},
		{"author_email", "john.doe@example.com"},
		{"message", "Test commit"},
		{"repository", "test-repo"},
		{"branch", "main"},
		{"pr_number", "123"},
		{"additions", "10"},
		{"deletions", "5"},
		{"files", "file1.go,file2.go"},
		{"unknown_key", ""},
	}

	for _, tc := range testCases {
		result := engine.getCommitValue(commit, tc.key)
		if result != tc.expected {
			t.Errorf("getCommitValue(commit, %s) = %s, want %s", tc.key, result, tc.expected)
		}
	}
}

func TestCorrelationEngine_FuzzyMatch(t *testing.T) {
	engine := &CorrelationEngine{}

	testCases := []struct {
		s1       string
		s2       string
		expected bool
	}{
		{"john", "john", true},
		{"john", "jon", true},
		{"john", "johny", true},
		{"john", "xyz", false},
		{"", "", true},
		{"a", "b", false},
	}

	for _, tc := range testCases {
		result := engine.fuzzyMatch(tc.s1, tc.s2)
		if result != tc.expected {
			t.Errorf("fuzzyMatch(%s, %s) = %v, want %v", tc.s1, tc.s2, result, tc.expected)
		}
	}
}
