package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfig holds all user‑adjustable settings.
type AppConfig struct {
	AutoSync                   bool   `json:"auto_sync"`
	CalendarID                 string `json:"calendar_id"` // if empty, primary is used
	NotificationThresholdHours int    `json:"notification_threshold_hours"`
}

func defaultConfig() AppConfig {
	return AppConfig{
		AutoSync:                   false,
		CalendarID:                 "",
		NotificationThresholdHours: 24,
	}
}

// GetConfigPath returns the full path to the config file.
func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".config", "ani-rem")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, "config.json")
}

// LoadConfig reads and parses the config file. If the file does not exist
// or is corrupted, the default config is returned.
func LoadConfig() (AppConfig, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultConfig(), nil
		}
		return defaultConfig(), fmt.Errorf("cannot read config: %w", err)
	}

	var cfg AppConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		// Corrupted file – return defaults but don't overwrite.
		return defaultConfig(), fmt.Errorf("config broken, using defaults: %w", err)
	}

	// Make sure threshold is reasonable
	if cfg.NotificationThresholdHours <= 0 {
		cfg.NotificationThresholdHours = 24
	}
	return cfg, nil
}

// SaveConfig writes the config to disk atomically.
func SaveConfig(cfg AppConfig) error {
	path := GetConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: temp file -> rename
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
