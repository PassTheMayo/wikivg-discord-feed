package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	WebhookURL    string        `yaml:"webhook_url"`
	CheckInterval time.Duration `yaml:"check_interval"`
}

func (c *Config) ReadFile(file string) error {
	data, err := os.ReadFile(file)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, c)
}
