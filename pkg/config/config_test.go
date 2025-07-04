package config

import (
	"testing"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/types"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.TimeWindow != 15*time.Minute {
		t.Errorf("Expected time window to be 15 minutes, got %v", config.TimeWindow)
	}

	if len(config.AttributeRules) != 1 {
		t.Errorf("Expected 1 attribute rule, got %d", len(config.AttributeRules))
	}

	if config.AttributeRules[0].EventKey != "user_id" {
		t.Errorf("Expected event key to be 'user_id', got %s", config.AttributeRules[0].EventKey)
	}

	if config.AttributeRules[0].CommitKey != "author_email" {
		t.Errorf("Expected commit key to be 'author_email', got %s", config.AttributeRules[0].CommitKey)
	}

	if config.AttributeRules[0].MatchType != types.EXACT {
		t.Errorf("Expected match type to be EXACT, got %v", config.AttributeRules[0].MatchType)
	}

	if !config.AttributeRules[0].Required {
		t.Errorf("Expected attribute rule to be required")
	}

	if config.ScoreWeights["temporal"] != 0.5 {
		t.Errorf("Expected temporal weight to be 0.5, got %f", config.ScoreWeights["temporal"])
	}

	if config.ScoreWeights["attribute"] != 0.5 {
		t.Errorf("Expected attribute weight to be 0.5, got %f", config.ScoreWeights["attribute"])
	}
}

func TestAIInferenceConfig(t *testing.T) {
	config := AIInferenceConfig()

	if config.TimeWindow != 15*time.Minute {
		t.Errorf("Expected time window to be 15 minutes, got %v", config.TimeWindow)
	}

	if len(config.AttributeRules) != 2 {
		t.Errorf("Expected 2 attribute rules, got %d", len(config.AttributeRules))
	}

	if config.AttributeRules[0].EventKey != "user_id" {
		t.Errorf("Expected first event key to be 'user_id', got %s", config.AttributeRules[0].EventKey)
	}

	if config.AttributeRules[1].EventKey != "project" {
		t.Errorf("Expected second event key to be 'project', got %s", config.AttributeRules[1].EventKey)
	}

	if config.AttributeRules[0].Required != true {
		t.Errorf("Expected first rule to be required")
	}

	if config.AttributeRules[1].Required != false {
		t.Errorf("Expected second rule to be optional")
	}

	if config.ScoreWeights["temporal"] != 0.6 {
		t.Errorf("Expected temporal weight to be 0.6, got %f", config.ScoreWeights["temporal"])
	}

	if config.ScoreWeights["attribute"] != 0.4 {
		t.Errorf("Expected attribute weight to be 0.4, got %f", config.ScoreWeights["attribute"])
	}
}

func TestDeploymentConfig(t *testing.T) {
	config := DeploymentConfig()

	if config.TimeWindow != 2*time.Hour {
		t.Errorf("Expected time window to be 2 hours, got %v", config.TimeWindow)
	}

	if len(config.AttributeRules) != 2 {
		t.Errorf("Expected 2 attribute rules, got %d", len(config.AttributeRules))
	}

	if config.AttributeRules[0].EventKey != "commit_sha" {
		t.Errorf("Expected first event key to be 'commit_sha', got %s", config.AttributeRules[0].EventKey)
	}

	if config.AttributeRules[0].CommitKey != "sha" {
		t.Errorf("Expected first commit key to be 'sha', got %s", config.AttributeRules[0].CommitKey)
	}

	if config.AttributeRules[0].MatchType != types.EXACT {
		t.Errorf("Expected first match type to be EXACT, got %v", config.AttributeRules[0].MatchType)
	}

	if config.AttributeRules[1].MatchType != types.REGEX {
		t.Errorf("Expected second match type to be REGEX, got %v", config.AttributeRules[1].MatchType)
	}

	if config.ScoreWeights["temporal"] != 0.3 {
		t.Errorf("Expected temporal weight to be 0.3, got %f", config.ScoreWeights["temporal"])
	}

	if config.ScoreWeights["attribute"] != 0.7 {
		t.Errorf("Expected attribute weight to be 0.7, got %f", config.ScoreWeights["attribute"])
	}
}

func TestConfigManager_NewConfigManager(t *testing.T) {
	configPath := "/test/path"
	cm := NewConfigManager(configPath)

	if cm.configPath != configPath {
		t.Errorf("Expected config path to be %s, got %s", configPath, cm.configPath)
	}
}
