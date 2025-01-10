// config/config.go
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Mattermost MattermostConfig `toml:"mattermost"`
	Feeds      FeedsConfig      `toml:"feeds"`
	Logging    LoggingConfig    `toml:"logging"`
	API        APIConfig        `toml:"api"`
}

type LoggingConfig struct {
	OutputToTerminal bool `toml:"output_to_terminal"`
}

type MattermostConfig struct {
	SecretURL string `toml:"secret_url"`
}

type FeedsConfig struct {
	URLs        []string `toml:"urls"`
	RescanDelay int      `toml:"rescan_delay"`
}

type APIConfig struct {
	Port int `toml:"port"`
}

func LoadConfig(filename string) (*Config, error) {
	data, readErr := os.ReadFile(filename)
	if readErr != nil {
		return nil, readErr
	}

	var config Config
	unmarshalErr := toml.Unmarshal(data, &config)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}

	validateErr := ValidateConfig(&config)
	if validateErr != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", validateErr)
	}

	return &config, nil
}

func FindValidConfigFiles() ([]string, error) {
	files, err := filepath.Glob("*.toml")
	if err != nil {
		return nil, err
	}

	var validFiles []string
	for _, file := range files {
		if isValidConfigFile(file) {
			validFiles = append(validFiles, file)
		}
	}

	return validFiles, nil
}

func isValidConfigFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		return strings.HasPrefix(line, "[mattermost]")
	}

	return false
}

func GetSingleConfigFile(configFlag string) (string, error) {
	if configFlag != "" {
		return configFlag, nil
	}

	configFiles, err := FindValidConfigFiles()
	if err != nil {
		return "", err
	}

	switch len(configFiles) {
	case 0:
		return "", fmt.Errorf("no valid config files found")
	case 1:
		return configFiles[0], nil
	default:
		return "", fmt.Errorf("multiple valid config files found: %v", configFiles)
	}
}