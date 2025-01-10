// config/validation.go
package config

import (
	"fmt"
	"net/url"
)

func ValidateConfig(cfg *Config) error {
	if err := validateMattermostConfig(&cfg.Mattermost); err != nil {
		return err
	}

	if err := validateFeedsConfig(&cfg.Feeds); err != nil {
		return err
	}

	if err := validateAPIConfig(&cfg.API); err != nil {
		return err
	}

	return nil
}

func validateMattermostConfig(cfg *MattermostConfig) error {
	if cfg.SecretURL == "" {
		return fmt.Errorf("mattermost.secret_url must be provided")
	}

	_, err := url.Parse(cfg.SecretURL)
	if err != nil {
		return fmt.Errorf("invalid mattermost.secret_url: %w", err)
	}

	return nil
}

func validateFeedsConfig(cfg *FeedsConfig) error {
	if len(cfg.URLs) == 0 {
		return fmt.Errorf("feeds.urls must contain at least one URL")
	}

	for _, feedURL := range cfg.URLs {
		_, err := url.Parse(feedURL)
		if err != nil {
			return fmt.Errorf("invalid feeds.url: %w", err)
		}
	}

	if cfg.RescanDelay <= 0 {
		return fmt.Errorf("feeds.rescan_delay must be greater than zero")
	}

	return nil
}

func validateAPIConfig(cfg *APIConfig) error {
	if cfg.Port <= 0 || cfg.Port > 65535 {
		return fmt.Errorf("api.port must be between 1 and 65535")
	}
	return nil
}
