package config

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	cfg, err := parseAndValidateConfigBytes([]byte(`
github:
  - repo: streetwave/gh-job
    workflow: action.yml

glue:
  - name: glue-job
`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := &Config{
		Github: []GithubJob{{Repo: "streetwave/gh-job", Workflow: "action.yml"}},
		Glue:   []GlueJob{{Name: "glue-job"}},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}

func TestParseConfigInvalid(t *testing.T) {
	cases := map[string]string{
		"repo without owner": "github:\n  - repo: gh-job\n    workflow: a.yml",
		"missing workflow":   "github:\n  - repo: a/b",
		"missing glue name":  "glue:\n  - name: \"\"",
		"github not array":   "github: a/b",
	}
	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := parseAndValidateConfigBytes([]byte(input)); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestParseConfigEmpty(t *testing.T) {
	if _, err := parseAndValidateConfigBytes(nil); !errors.Is(err, ErrNoJobsInConfig) {
		t.Errorf("got %v, want %v", err, ErrNoJobsInConfig)
	}
}

func TestLoadConfigFromPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jobs.yaml")
	if err := os.WriteFile(path, []byte("glue:\n  - name: job"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Glue) != 1 || cfg.Glue[0].Name != "job" {
		t.Errorf("unexpected config %+v", cfg)
	}
}

func TestLoadConfigNotFound(t *testing.T) {
	t.Chdir(t.TempDir())
	if _, err := LoadConfig("does-not-exist.yaml"); !errors.Is(err, ErrConfigFileNotFound) {
		t.Errorf("got %v, want %v", err, ErrConfigFileNotFound)
	}
}
