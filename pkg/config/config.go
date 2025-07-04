package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fraser-isbester/git-snap/pkg/types"
	"github.com/spf13/viper"
)

type ConfigManager struct {
	configPath string
}

func NewConfigManager(configPath string) *ConfigManager {
	return &ConfigManager{configPath: configPath}
}

func (cm *ConfigManager) LoadConfig(name string) (*types.SnapConfig, error) {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType("yaml")
	v.AddConfigPath(cm.configPath)
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config types.SnapConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Parse duration string if needed
	if v.IsSet("time_window") {
		timeWindowStr := v.GetString("time_window")
		if duration, err := time.ParseDuration(timeWindowStr); err == nil {
			config.TimeWindow = duration
		}
	}

	return &config, nil
}

func (cm *ConfigManager) SaveConfig(name string, config *types.SnapConfig) error {
	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType("yaml")
	v.AddConfigPath(cm.configPath)

	v.Set("time_window", config.TimeWindow.String())
	v.Set("attribute_rules", config.AttributeRules)
	v.Set("score_weights", config.ScoreWeights)

	configFile := filepath.Join(cm.configPath, name+".yaml")
	return v.WriteConfigAs(configFile)
}

func (cm *ConfigManager) ListConfigs() ([]string, error) {
	files, err := os.ReadDir(cm.configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	var configs []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".yaml") {
			name := strings.TrimSuffix(file.Name(), ".yaml")
			configs = append(configs, name)
		}
	}

	return configs, nil
}

func (cm *ConfigManager) EnsureConfigDirectory() error {
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		return os.MkdirAll(cm.configPath, 0755)
	}
	return nil
}

func DefaultConfig() *types.SnapConfig {
	return &types.SnapConfig{
		TimeWindow: 15 * time.Minute,
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
}

func AIInferenceConfig() *types.SnapConfig {
	return &types.SnapConfig{
		TimeWindow: 15 * time.Minute,
		AttributeRules: []types.AttributeRule{
			{
				EventKey:  "user_id",
				CommitKey: "author_email",
				MatchType: types.EXACT,
				Required:  true,
			},
			{
				EventKey:  "project",
				CommitKey: "repository",
				MatchType: types.CONTAINS,
				Required:  false,
			},
		},
		ScoreWeights: map[string]float64{
			"temporal":  0.6,
			"attribute": 0.4,
		},
	}
}

func DeploymentConfig() *types.SnapConfig {
	return &types.SnapConfig{
		TimeWindow: 2 * time.Hour,
		AttributeRules: []types.AttributeRule{
			{
				EventKey:  "commit_sha",
				CommitKey: "sha",
				MatchType: types.EXACT,
				Required:  true,
			},
			{
				EventKey:  "environment",
				CommitKey: "branch",
				MatchType: types.REGEX,
				Required:  false,
			},
		},
		ScoreWeights: map[string]float64{
			"temporal":  0.3,
			"attribute": 0.7,
		},
	}
}

func (cm *ConfigManager) CreateDefaultConfigs() error {
	if err := cm.EnsureConfigDirectory(); err != nil {
		return err
	}

	configs := map[string]*types.SnapConfig{
		"default":      DefaultConfig(),
		"ai-inference": AIInferenceConfig(),
		"deployment":   DeploymentConfig(),
	}

	for name, config := range configs {
		if err := cm.SaveConfig(name, config); err != nil {
			return fmt.Errorf("failed to create config %s: %w", name, err)
		}
	}

	return nil
}

func init() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}
