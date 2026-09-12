package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	vekconfig "github.com/vekio/config"
	appconfig "github.com/vekio/cspeek/internal/config"
	"github.com/vekio/cspeek/internal/matches"
)

func TestRootProvidesConfigurationCommandsAndFlag(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	command, err := New()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	command.Writer = &output
	if err := command.Run(context.Background(), []string{"cspeek", "--help"}); err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"config", "live", "upcoming", "--config"} {
		if !strings.Contains(output.String(), expected) {
			t.Errorf("help does not contain %q:\n%s", expected, output.String())
		}
	}
}

func TestConfigInitWritesAPIKeySetting(t *testing.T) {
	file, err := vekconfig.NewYAMLConfigFile[appconfig.Config]("cspeek-test", "config.yml")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "custom.yml")
	command := newRootCommand(file, appconfig.Default(), func(appconfig.Config) (matches.Source, error) {
		t.Fatal("config command created an API client")
		return nil, nil
	})
	if err := command.Run(context.Background(), []string{"cspeek", "--config", path, "config", "init"}); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(content); !strings.Contains(got, "api_key: "+appconfig.Default().APIKey) {
		t.Fatalf("configuration = %q", got)
	}
}
