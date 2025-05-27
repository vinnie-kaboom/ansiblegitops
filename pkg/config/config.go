package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

// Config holds the application configuration
type Config struct {
	Git struct {
		URL          string `yaml:"url"`
		Branch       string `yaml:"branch"`
		PollInterval int    `yaml:"poll_interval"`
		Auth         struct {
			Type     string `yaml:"type"`
			Username string `yaml:"username"`
			Token    string `yaml:"token"`
		} `yaml:"auth"`
	} `yaml:"git"`
	Ansible struct {
		PlaybookDir   string `yaml:"playbook_dir"`
		InventoryFile string `yaml:"inventory_file"`
	} `yaml:"ansible"`
}

// Load reads configuration from config.yaml and processes environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	// Replace environment variables in the yaml content
	content := string(data)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			placeholder := "${" + parts[0] + "}"
			content = strings.Replace(content, placeholder, parts[1], -1)
		}
	}

	err = yaml.Unmarshal([]byte(content), cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}