package config

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type ElasticsearchConfig struct {
	URL      string
	Username string
	Password string
}

// LoadElasticsearchConfig reads the private KEY=value file created in Lesson 31.
// Environment variables take precedence over file values. ES_CONFIG_FILE can
// select a different file; otherwise the user's config directory is used.
func LoadElasticsearchConfig() (ElasticsearchConfig, error) {
	path := os.Getenv("ES_CONFIG_FILE")
	explicit := path != ""
	if !explicit {
		dir, err := os.UserConfigDir()
		if err != nil {
			return ElasticsearchConfig{}, fmt.Errorf("locate Elasticsearch configuration: %w", err)
		}
		path = filepath.Join(dir, "networkai", "elasticsearch.env")
	}
	values := make(map[string]string)
	file, err := os.Open(path)
	if err != nil {
		if explicit || !os.IsNotExist(err) {
			return ElasticsearchConfig{}, fmt.Errorf("open Elasticsearch configuration: %w", err)
		}
	} else {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for lineNumber := 1; scanner.Scan(); lineNumber++ {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				return ElasticsearchConfig{}, fmt.Errorf("invalid Elasticsearch configuration at line %d: expected KEY=value", lineNumber)
			}
			value = strings.TrimSpace(value)
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			values[strings.TrimSpace(key)] = value
		}
		if err := scanner.Err(); err != nil {
			return ElasticsearchConfig{}, fmt.Errorf("read Elasticsearch configuration: %w", err)
		}
	}
	for _, key := range []string{"ES_URL", "ES_USERNAME", "ES_PASSWORD"} {
		if value, ok := os.LookupEnv(key); ok {
			values[key] = value
		}
		if values[key] == "" {
			return ElasticsearchConfig{}, fmt.Errorf("%s is required; set it in the environment or %s", key, path)
		}
	}
	address, err := url.Parse(values["ES_URL"])
	if err != nil || (address.Scheme != "http" && address.Scheme != "https") || address.Host == "" || address.User != nil {
		return ElasticsearchConfig{}, fmt.Errorf("ES_URL must be an HTTP(S) URL without embedded credentials")
	}
	return ElasticsearchConfig{
		URL: values["ES_URL"], Username: values["ES_USERNAME"], Password: values["ES_PASSWORD"],
	}, nil
}
