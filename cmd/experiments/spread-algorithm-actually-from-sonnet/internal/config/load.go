package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	if cfg.Version == "" {
		cfg.Version = "0.1"
	}
	return &cfg, nil
}

// ResolveFromFile loads then resolves spreads relative to config directory.
func ResolveFromFile(path string) ([]SpreadSpec, error) {
	cfg, err := Load(path)
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Dir(path)
	return cfg.Resolve(baseDir)
}
