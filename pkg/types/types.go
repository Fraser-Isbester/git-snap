package types

import "time"

type SnapEvent struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Attributes map[string]interface{} `json:"attributes"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type EnrichedCommit struct {
	SHA         string    `json:"sha"`
	Author      string    `json:"author"`
	AuthorEmail string    `json:"author_email"`
	Committer   string    `json:"committer"`
	Timestamp   time.Time `json:"timestamp"`
	Message     string    `json:"message"`
	Files       []string  `json:"files"`
	Branch      string    `json:"branch"`
	Repository  string    `json:"repository"`
	PRNumber    *int      `json:"pr_number,omitempty"`
	Additions   int       `json:"additions"`
	Deletions   int       `json:"deletions"`
	Parents     []string  `json:"parents"`
}

type SnapConfig struct {
	TimeWindow     time.Duration      `yaml:"time_window"`
	AttributeRules []AttributeRule    `yaml:"attribute_rules"`
	ScoreWeights   map[string]float64 `yaml:"score_weights"`
}

type AttributeRule struct {
	EventKey  string    `yaml:"event_key"`
	CommitKey string    `yaml:"commit_key"`
	MatchType MatchType `yaml:"match_type"`
	Required  bool      `yaml:"required"`
}

type MatchType int

const (
	EXACT MatchType = iota
	CONTAINS
	REGEX
	FUZZY
)

func (m MatchType) String() string {
	switch m {
	case EXACT:
		return "exact"
	case CONTAINS:
		return "contains"
	case REGEX:
		return "regex"
	case FUZZY:
		return "fuzzy"
	default:
		return "unknown"
	}
}

func (m MatchType) MarshalYAML() (interface{}, error) {
	return m.String(), nil
}

func (m *MatchType) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}

	switch s {
	case "exact":
		*m = EXACT
	case "contains":
		*m = CONTAINS
	case "regex":
		*m = REGEX
	case "fuzzy":
		*m = FUZZY
	default:
		*m = EXACT
	}
	return nil
}

type CorrelationResult struct {
	Event     SnapEvent       `json:"event"`
	Commit    EnrichedCommit  `json:"commit"`
	Score     float64         `json:"score"`
	Matches   map[string]bool `json:"matches"`
	TimeDelta time.Duration   `json:"time_delta"`
}
