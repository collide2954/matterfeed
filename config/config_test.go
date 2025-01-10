// config/config_test.go
package config_test

import (
	"os"
	"testing"

	"matterfeed/config"
)

func TestLoadValidConfig(t *testing.T) {
	cfg, cleanup, err := setupTestConfig(`
[mattermost]
secret_url = "https://example.com/hooks/abcdefg123456"

[feeds]
urls = ["http://example.com/feed"]
rescan_delay = 300

[logging]
output_to_terminal = true

[api]
port = 8080
`)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if cfg.Mattermost.SecretURL != "https://example.com/hooks/abcdefg123456" {
		t.Errorf("Expected Mattermost.SecretURL to be 'https://example.com/hooks/abcdefg123456', got '%s'", cfg.Mattermost.SecretURL)
	}
	if len(cfg.Feeds.URLs) != 1 || cfg.Feeds.URLs[0] != "http://example.com/feed" {
		t.Errorf("Expected Feeds.URLs to be ['http://example.com/feed'], got %v", cfg.Feeds.URLs)
	}
	if cfg.Feeds.RescanDelay != 300 {
		t.Errorf("Expected Feeds.RescanDelay to be 300, got %d", cfg.Feeds.RescanDelay)
	}
	if !cfg.Logging.OutputToTerminal {
		t.Errorf("Expected Logging.OutputToTerminal to be true, got false")
	}
	if cfg.API.Port != 8080 {
		t.Errorf("Expected API.Port to be 8080, got %d", cfg.API.Port)
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	_, cleanup, err := setupTestConfig(`
[mattermost]
# secret_url is missing

[feeds]
urls = ["http://example.com/feed"]
rescan_delay = 300

[logging]
output_to_terminal = true

[api]
port = 8080
`)
	if err == nil {
		t.Errorf("Expected an error due to missing secret_url, got none")
	}
	if cleanup != nil {
		cleanup()
	}

	_, err = config.LoadConfig("non_existent_file.toml")
	if err == nil {
		t.Errorf("Expected an error when loading non-existent file, got none")
	}
}

func TestValidateConfig(t *testing.T) {
	cfg := config.Config{
		Mattermost: config.MattermostConfig{
			SecretURL: "https://example.com/hooks/abcdefg123456",
		},
		Feeds: config.FeedsConfig{
			URLs:        []string{"http://example.com/feed"},
			RescanDelay: 300,
		},
		Logging: config.LoggingConfig{
			OutputToTerminal: true,
		},
		API: config.APIConfig{
			Port: 8080,
		},
	}

	if err := config.ValidateConfig(&cfg); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.Config
	}{
		{
			name: "Missing Mattermost SecretURL",
			cfg: config.Config{
				Mattermost: config.MattermostConfig{},
				Feeds: config.FeedsConfig{
					URLs:        []string{"http://example.com/feed"},
					RescanDelay: 300,
				},
				Logging: config.LoggingConfig{
					OutputToTerminal: true,
				},
				API: config.APIConfig{
					Port: 8080,
				},
			},
		},
		{
			name: "Empty Feeds URLs",
			cfg: config.Config{
				Mattermost: config.MattermostConfig{
					SecretURL: "https://example.com/hooks/abcdefg123456",
				},
				Feeds: config.FeedsConfig{
					URLs:        []string{},
					RescanDelay: 300,
				},
				Logging: config.LoggingConfig{
					OutputToTerminal: true,
				},
				API: config.APIConfig{
					Port: 8080,
				},
			},
		},
		{
			name: "Zero Rescan Delay",
			cfg: config.Config{
				Mattermost: config.MattermostConfig{
					SecretURL: "https://example.com/hooks/abcdefg123456",
				},
				Feeds: config.FeedsConfig{
					URLs:        []string{"http://example.com/feed"},
					RescanDelay: 0,
				},
				Logging: config.LoggingConfig{
					OutputToTerminal: true,
				},
				API: config.APIConfig{
					Port: 8080,
				},
			},
		},
		{
			name: "Invalid API Port",
			cfg: config.Config{
				Mattermost: config.MattermostConfig{
					SecretURL: "https://example.com/hooks/abcdefg123456",
				},
				Feeds: config.FeedsConfig{
					URLs:        []string{"http://example.com/feed"},
					RescanDelay: 300,
				},
				Logging: config.LoggingConfig{
					OutputToTerminal: true,
				},
				API: config.APIConfig{
					Port: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := config.ValidateConfig(&tt.cfg); err == nil {
				t.Errorf("Expected an error, got none")
			}
		})
	}
}

func setupTestConfig(content string) (*config.Config, func(), error) {
	tmpFile, err := os.CreateTemp("", "test_config_*.toml")
	if err != nil {
		return nil, nil, err
	}

	_, err = tmpFile.WriteString(content)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return nil, nil, err
	}

	err = tmpFile.Close()
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, nil, err
	}

	cfg, err := config.LoadConfig(tmpFile.Name())
	if err != nil {
		os.Remove(tmpFile.Name())
		return nil, nil, err
	}

	cleanup := func() {
		os.Remove(tmpFile.Name())
	}

	return cfg, cleanup, nil
}
