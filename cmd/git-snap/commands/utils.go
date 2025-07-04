package commands

import (
	"os"
	"path/filepath"
)

func getConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./config"
	}
	return filepath.Join(home, ".git-snap", "config")
}
