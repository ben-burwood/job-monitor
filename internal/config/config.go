package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// DefaultConfigurationFilePath is the default path that will be used to search for the configuration file
	// if a custom path isn't configured through the JOB_MONITOR_CONFIG_PATH environment variable
	DefaultConfigurationFilePath = "config/config.toml"
)

var (
	// ErrConfigFileNotFound is an error returned when a configuration file could not be found
	ErrConfigFileNotFound = errors.New("configuration file not found")

	// ErrNoJobsInConfig is an error returned when no jobs are configured in the configuration file
	ErrNoJobsInConfig = errors.New("no jobs configured in configuration file")
)

// GithubJob is a GitHub Actions workflow to track, run against the repository's default branch
type GithubJob struct {
	Repo     string `toml:"repo"`     // owner/name
	Workflow string `toml:"workflow"` // workflow filename in .github/workflows/
}

// GlueJob is an AWS Glue job to track
type GlueJob struct {
	Name string `toml:"name"`
}

type Config struct {
	Github []GithubJob `toml:"github"`
	Glue   []GlueJob   `toml:"glue"`
}

// LoadConfig loads the configuration from the specified path, falling back to DefaultConfigurationFilePath
func LoadConfig(configPath string) (*Config, error) {
	for _, path := range []string{configPath, DefaultConfigurationFilePath} {
		if len(path) == 0 {
			continue
		}
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("error reading configuration from %s: %w", path, err)
		}
		log.Printf("[config.LoadConfig] Reading configuration from %s", path)
		return parseAndValidateConfigBytes(data)
	}
	return nil, ErrConfigFileNotFound
}

// parseAndValidateConfigBytes parses a TOML configuration into a Config struct and validates its parameters
func parseAndValidateConfigBytes(data []byte) (*Config, error) {
	var config Config
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	for i, job := range config.Github {
		if !strings.Contains(job.Repo, "/") {
			return nil, fmt.Errorf("github[%d].repo must be 'owner/name'", i)
		}
		if len(job.Workflow) == 0 {
			return nil, fmt.Errorf("github[%d].workflow must be a non-empty filename", i)
		}
	}
	for i, job := range config.Glue {
		if len(job.Name) == 0 {
			return nil, fmt.Errorf("glue[%d].name must be a non-empty string", i)
		}
	}

	if len(config.Github) == 0 && len(config.Glue) == 0 {
		return nil, ErrNoJobsInConfig
	}
	return &config, nil
}
