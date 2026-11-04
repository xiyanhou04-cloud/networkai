package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func isolateConfig(t *testing.T) string {
	t.Helper()
	for _, key := range []string{"ES_URL", "ES_USERNAME", "ES_PASSWORD", "ES_CONFIG_FILE"} {
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path := filepath.Join(dir, "networkai", "elasticsearch.env")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestPrivateConfigAndEnvironmentPrecedence(t *testing.T) {
	path := isolateConfig(t)
	if err := os.WriteFile(path, []byte("# private settings\nES_URL=http://localhost:9200\nES_USERNAME='course'\nES_PASSWORD=\"file=password\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadElasticsearchConfig()
	if err != nil || cfg.URL != "http://localhost:9200" || cfg.Username != "course" || cfg.Password != "file=password" {
		t.Fatal("private configuration was not loaded correctly")
	}
	t.Setenv("ES_PASSWORD", "environment-password")
	cfg, err = LoadElasticsearchConfig()
	if err != nil || cfg.Password != "environment-password" {
		t.Fatal("environment did not override the private file")
	}
	t.Setenv("ES_PASSWORD", "")
	if _, err := LoadElasticsearchConfig(); err == nil || !strings.Contains(err.Error(), "ES_PASSWORD") {
		t.Fatal("explicitly empty password should fail validation")
	}
}

func TestEnvironmentOnlyConfiguration(t *testing.T) {
	isolateConfig(t)
	t.Setenv("ES_URL", "http://localhost:9200")
	t.Setenv("ES_USERNAME", "course")
	t.Setenv("ES_PASSWORD", "private-password")
	if _, err := LoadElasticsearchConfig(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ES_CONFIG_FILE", filepath.Join(t.TempDir(), "missing.env"))
	if _, err := LoadElasticsearchConfig(); err == nil {
		t.Fatal("an explicitly selected missing file should fail")
	}
}

func TestConfigurationErrorsDoNotExposeSecrets(t *testing.T) {
	path := isolateConfig(t)
	secret := "do-not-expose-this-password"
	t.Setenv("ES_URL", "http://course:"+secret+"@localhost:9200")
	t.Setenv("ES_USERNAME", "course")
	t.Setenv("ES_PASSWORD", secret)
	if _, err := LoadElasticsearchConfig(); err == nil || strings.Contains(err.Error(), secret) {
		t.Fatal("URL validation should fail without exposing embedded credentials")
	}
	if err := os.WriteFile(path, []byte(secret), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadElasticsearchConfig(); err == nil || strings.Contains(err.Error(), secret) {
		t.Fatal("malformed file errors should not expose file content")
	}
}
