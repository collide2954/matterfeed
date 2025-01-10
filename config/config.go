// config/config.go
package config

import (
	"bufio"
	"errors"
	"fmt"
	"log"
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

func LoadConfig(configFlag string) (*Config, error) {
	var filename string
	if configFlag != "" {
		filename = configFlag
	} else {
		configFiles, err := FindValidConfigFiles()
		if err != nil {
			return nil, err
		}
		switch len(configFiles) {
		case 0:
			return nil, errors.New("no valid config files found")
		case 1:
			filename = configFiles[0]
		default:
			return nil, fmt.Errorf("multiple valid config files found: %v", configFiles)
		}
	}

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
		log.Printf("Error opening file %s: %v", filename, err)
		return false
	}
	defer func(file *os.File) {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("Error closing file %s: %v", filename, closeErr)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		line := scanner.Text()
		return strings.HasPrefix(line, "[mattermost]")
	}

	return false
}
