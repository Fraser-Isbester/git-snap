package correlation

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/types"
)

type CorrelationEngine struct {
	config types.SnapConfig
}

func NewCorrelationEngine(config types.SnapConfig) *CorrelationEngine {
	return &CorrelationEngine{config: config}
}

func (e *CorrelationEngine) SnapToCommits(events []types.SnapEvent, commits []types.EnrichedCommit) []types.CorrelationResult {
	var results []types.CorrelationResult

	for _, event := range events {
		for _, commit := range commits {
			if !e.isWithinTimeWindow(event, commit) {
				continue
			}

			matches, score := e.calculateCorrelation(event, commit)
			if score > 0 {
				results = append(results, types.CorrelationResult{
					Event:     event,
					Commit:    commit,
					Score:     score,
					Matches:   matches,
					TimeDelta: e.calculateTimeDelta(event, commit),
				})
			}
		}
	}

	return e.sortAndFilterResults(results)
}

func (e *CorrelationEngine) isWithinTimeWindow(event types.SnapEvent, commit types.EnrichedCommit) bool {
	timeDelta := e.calculateTimeDelta(event, commit)
	return timeDelta <= e.config.TimeWindow
}

func (e *CorrelationEngine) calculateTimeDelta(event types.SnapEvent, commit types.EnrichedCommit) time.Duration {
	if event.Timestamp.After(commit.Timestamp) {
		return event.Timestamp.Sub(commit.Timestamp)
	}
	return commit.Timestamp.Sub(event.Timestamp)
}

func (e *CorrelationEngine) calculateCorrelation(event types.SnapEvent, commit types.EnrichedCommit) (map[string]bool, float64) {
	matches := make(map[string]bool)
	requiredMatches := 0
	totalRequired := 0
	optionalMatches := 0
	totalOptional := 0

	for _, rule := range e.config.AttributeRules {
		if rule.Required {
			totalRequired++
		} else {
			totalOptional++
		}

		eventValue := e.getEventValue(event, rule.EventKey)
		commitValue := e.getCommitValue(commit, rule.CommitKey)

		if e.matchValues(eventValue, commitValue, rule.MatchType) {
			matches[rule.EventKey] = true
			if rule.Required {
				requiredMatches++
			} else {
				optionalMatches++
			}
		} else {
			matches[rule.EventKey] = false
		}
	}

	if totalRequired > 0 && requiredMatches < totalRequired {
		return matches, 0
	}

	attributeScore := 1.0
	if totalRequired > 0 {
		attributeScore = float64(requiredMatches) / float64(totalRequired)
	}

	if totalOptional > 0 {
		optionalScore := float64(optionalMatches) / float64(totalOptional)
		attributeScore = (attributeScore + optionalScore) / 2
	}

	temporalScore := e.calculateTemporalScore(event, commit)

	temporalWeight := e.config.ScoreWeights["temporal"]
	if temporalWeight == 0 {
		temporalWeight = 0.5
	}

	attributeWeight := e.config.ScoreWeights["attribute"]
	if attributeWeight == 0 {
		attributeWeight = 0.5
	}

	totalScore := (temporalScore * temporalWeight) + (attributeScore * attributeWeight)
	return matches, totalScore
}

func (e *CorrelationEngine) calculateTemporalScore(event types.SnapEvent, commit types.EnrichedCommit) float64 {
	timeDelta := e.calculateTimeDelta(event, commit)
	return math.Max(0, 1.0-(float64(timeDelta)/float64(e.config.TimeWindow)))
}

func (e *CorrelationEngine) getEventValue(event types.SnapEvent, key string) string {
	if value, exists := event.Attributes[key]; exists {
		return fmt.Sprintf("%v", value)
	}
	return ""
}

func (e *CorrelationEngine) getCommitValue(commit types.EnrichedCommit, key string) string {
	switch key {
	case "sha":
		return commit.SHA
	case "author":
		return commit.Author
	case "author_email":
		return commit.AuthorEmail
	case "committer":
		return commit.Committer
	case "message":
		return commit.Message
	case "repository":
		return commit.Repository
	case "branch":
		return commit.Branch
	case "pr_number":
		if commit.PRNumber != nil {
			return strconv.Itoa(*commit.PRNumber)
		}
		return ""
	case "additions":
		return strconv.Itoa(commit.Additions)
	case "deletions":
		return strconv.Itoa(commit.Deletions)
	case "files":
		return strings.Join(commit.Files, ",")
	default:
		return ""
	}
}

func (e *CorrelationEngine) matchValues(eventValue, commitValue string, matchType types.MatchType) bool {
	if eventValue == "" || commitValue == "" {
		return false
	}

	switch matchType {
	case types.EXACT:
		return eventValue == commitValue
	case types.CONTAINS:
		return strings.Contains(strings.ToLower(commitValue), strings.ToLower(eventValue))
	case types.REGEX:
		matched, err := regexp.MatchString(eventValue, commitValue)
		return err == nil && matched
	case types.FUZZY:
		return e.fuzzyMatch(eventValue, commitValue)
	default:
		return false
	}
}

func (e *CorrelationEngine) fuzzyMatch(s1, s2 string) bool {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	distance := e.levenshteinDistance(s1, s2)
	maxLen := float64(max(len(s1), len(s2)))

	if maxLen == 0 {
		return true
	}

	similarity := 1.0 - (float64(distance) / maxLen)
	return similarity >= 0.7
}

func (e *CorrelationEngine) levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1),
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func (e *CorrelationEngine) sortAndFilterResults(results []types.CorrelationResult) []types.CorrelationResult {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
