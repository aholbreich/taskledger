package repo

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds values read from .taskledger/config.yaml.
type Config struct {
	DefaultClaimTTL string `yaml:"default_claim_ttl"`
}

// LoadConfig reads the config file under ledger. Missing or unparseable values
// fall back to safe defaults (60m TTL).
func LoadConfig(ledger string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(ledger, ConfigFile))
	if err != nil {
		return nil, err
	}
	cfg := &Config{DefaultClaimTTL: "60m"}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.DefaultClaimTTL == "" {
		cfg.DefaultClaimTTL = "60m"
	}
	return cfg, nil
}
