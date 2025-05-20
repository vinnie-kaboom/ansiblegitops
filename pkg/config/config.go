package config

import (
	"github.com/spf13/viper"
	"os"
	"strings"
)

// GitAuth holds Git authentication configuration
type GitAuth struct {
	Type     string `mapstructure:"type"`
	Username string `mapstructure:"username"`
	Token    string `mapstructure:"token"`
	SSHKey   string `mapstructure:"ssh_key"`
}

// GitConfig holds Git repository configuration settings
type GitConfig struct {
	URL          string  `mapstructure:"url"`
	Branch       string  `mapstructure:"branch"`
	PollInterval int     `mapstructure:"poll_interval"`
	Auth         GitAuth `mapstructure:"auth"`
}

// AnsibleConfig holds Ansible execution configuration settings
type AnsibleConfig struct {
	PlaybookDir   string `mapstructure:"playbook_dir"`
	InventoryFile string `mapstructure:"inventory_file"`
}

// Config represents the main application configuration
type Config struct {
	Git     GitConfig     `mapstructure:"git"`
	Ansible AnsibleConfig `mapstructure:"ansible"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Set environment variable configuration
	v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Allow environment variables to be read
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Enable environment variable substitution
	for _, key := range []string{
		"git.url",
		"git.branch",
		"git.auth.username",
		"git.auth.token",
		"ansible.playbook_dir",
		"ansible.inventory_file",
	} {
		if val := v.GetString(key); strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
			envVar := strings.TrimSuffix(strings.TrimPrefix(val, "${"), "}")
			if envVal := os.Getenv(envVar); envVal != "" {
				v.Set(key, envVal)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
