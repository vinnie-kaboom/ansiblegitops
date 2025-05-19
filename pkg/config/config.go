package config

import (
    "github.com/spf13/viper"
    "log"
)

type Config struct {
    Git struct {
        URL          string `mapstructure:"url"`
        Branch       string `mapstructure:"branch"`
        PollInterval int    `mapstructure:"poll_interval"`
    } `mapstructure:"git"`
    Ansible struct {
        PlaybookDir   string `mapstructure:"playbook_dir"`
        InventoryFile string `mapstructure:"inventory_file"`
    } `mapstructure:"ansible"`
}

func LoadConfig(path string) (*Config, error) {
    v := viper.New()
    v.SetConfigFile(path)
    v.SetConfigType("yaml")
    if err := v.ReadInConfig(); err != nil {
        return nil, err
    }
    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}
