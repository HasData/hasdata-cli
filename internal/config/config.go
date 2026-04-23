package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	APIKey           string    `yaml:"api_key,omitempty"`
	Endpoint         string    `yaml:"endpoint,omitempty"`
	CheckUpdates     *bool     `yaml:"check_updates,omitempty"`
	LastUpdateCheck  time.Time `yaml:"last_update_check,omitempty"`
	LatestKnownVer   string    `yaml:"latest_known_version,omitempty"`
}

const DefaultEndpoint = "https://api.hasdata.com"

func Path() (string, error) {
	if p := os.Getenv("HASDATA_CONFIG"); p != "" {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hasdata", "config.yaml"), nil
}

func Load(path string) (*Config, error) {
	if path == "" {
		p, err := Path()
		if err != nil {
			return nil, err
		}
		path = p
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &c, nil
}

func Save(path string, c *Config) error {
	if path == "" {
		p, err := Path()
		if err != nil {
			return err
		}
		path = p
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// ResolveAPIKey applies precedence: flag > env > config.
func ResolveAPIKey(flagVal string, c *Config) string {
	if flagVal != "" {
		return flagVal
	}
	if v := os.Getenv("HASDATA_API_KEY"); v != "" {
		return v
	}
	if c != nil {
		return c.APIKey
	}
	return ""
}

func ResolveEndpoint(flagVal string, c *Config) string {
	if flagVal != "" {
		return flagVal
	}
	if v := os.Getenv("HASDATA_ENDPOINT"); v != "" {
		return v
	}
	if c != nil && c.Endpoint != "" {
		return c.Endpoint
	}
	return DefaultEndpoint
}

func ShouldCheckUpdates(c *Config) bool {
	if c == nil || c.CheckUpdates == nil {
		return true
	}
	return *c.CheckUpdates
}
