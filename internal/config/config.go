// Package config defines CSpeek's local configuration.
package config

import (
	"fmt"
	"strings"

	vekconfig "github.com/vekio/config"
)

const (
	applicationName = "cspeek"
	fileName        = "config.yml"
	defaultAPIKey   = "replace-with-your-liquipedia-api-key"
)

// Config contains the credentials used to access Liquipedia.
type Config struct {
	APIKey string `json:"api_key" yaml:"api_key"`
}

// Default returns a configuration that can be created before an API key is set.
func Default() Config {
	return Config{APIKey: defaultAPIKey}
}

// Validate rejects values that are easy to copy incorrectly.
func (config Config) Validate() error {
	if config.APIKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}
	if strings.TrimSpace(config.APIKey) != config.APIKey {
		return fmt.Errorf("API key cannot contain surrounding whitespace")
	}
	return nil
}

// New creates the conventional CSpeek YAML configuration file.
func New() (*vekconfig.ConfigFile[Config], error) {
	return vekconfig.NewYAMLConfigFile[Config](applicationName, fileName)
}

var _ vekconfig.Validatable = Config{}
