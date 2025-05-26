package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	BotName  string `yaml:"bot_name"`
	Channel  string `yaml:"channel"`
	DataPath string `yaml:"data_path"`
	Commands struct {
		Queue struct {
			MaxSize         int `yaml:"max_size"`
			DefaultPosition int `yaml:"default_position"`
			DefaultPopCount int `yaml:"default_pop_count"`
		} `yaml:"queue"`
		Cooldowns struct {
			Default   int `yaml:"default"`
			Moderator int `yaml:"moderator"`
			VIP       int `yaml:"vip"`
		} `yaml:"cooldowns"`
	} `yaml:"commands"`
}

// Load loads the configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate required fields
	if config.Channel == "" {
		return nil, fmt.Errorf("channel is required in config")
	}
	if config.BotName == "" {
		return nil, fmt.Errorf("bot_name is required in config")
	}

	// Set default data path if not specified
	if config.DataPath == "" {
		config.DataPath = fmt.Sprintf("/app/data/%s", config.Channel)
	}

	// Set default command values if not specified
	if config.Commands.Queue.MaxSize == 0 {
		config.Commands.Queue.MaxSize = 100
	}
	if config.Commands.Queue.DefaultPosition == 0 {
		config.Commands.Queue.DefaultPosition = 1
	}
	if config.Commands.Queue.DefaultPopCount == 0 {
		config.Commands.Queue.DefaultPopCount = 1
	}
	if config.Commands.Cooldowns.Default == 0 {
		config.Commands.Cooldowns.Default = 5
	}
	if config.Commands.Cooldowns.Moderator == 0 {
		config.Commands.Cooldowns.Moderator = 2
	}
	if config.Commands.Cooldowns.VIP == 0 {
		config.Commands.Cooldowns.VIP = 3
	}

	return &config, nil
}
