package config

import (
	"gopkg.in/yaml.v2"
	"os"
	"strings"
)

// Config is a structure that holds the application's configuration settings.
// It contains nested structures for Git and Ansible configurations, including
// authentication details, repository settings, and playbook locations.
// All fields are mapped to corresponding YAML tags for configuration file
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

// Load reads and parses the configuration from config.yaml file.
// It processes environment variables in the configuration content,
// replacing ${VAR_NAME} placeholders with their actual values.
// Returns a pointer to the Config structure and any error encountered
// during loading or parsing.
func Load() (*Config, error) {
	// 1. Initialize empty configuration
	cfg := &Config{}

	// 2. Read configuration file
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	// 3. Process environment variables
	content := string(data)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			// 4. Replace environment variables placeholders
			placeholder := "${" + parts[0] + "}"
			content = strings.Replace(content, placeholder, parts[1], -1)
		}
	}

	// 5. Parse YAML content into configuration structure
	err = yaml.Unmarshal([]byte(content), cfg)
	if err != nil {
		return nil, err
	}

	// 6. Return parsed configuration
	return cfg, nil
}
