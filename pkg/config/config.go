package config

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Git struct {
		URL          string `yaml:"url" validate:"required"`
		Branch       string `yaml:"branch" validate:"required"`
		PollInterval int    `yaml:"poll_interval" validate:"gt=0"`
		Auth         struct {
			Type     string `yaml:"type" validate:"required"`
			Username string `yaml:"username" validate:"required"`
			Token    string `yaml:"token" validate:"required"`
		} `yaml:"auth"`
	} `yaml:"git"`
	Ansible struct {
		PlaybookDir   string `yaml:"playbook_dir" validate:"required"`
		InventoryFile string `yaml:"inventory_file" validate:"required"`
	} `yaml:"ansible"`
}

// Validate checks if all required configuration values are set
func (c *Config) Validate() error {
	var errMsgs []string

	// Helper function to validate a value based on its validation tag
	validateField := func(value reflect.Value, field reflect.StructField, path string) {
		tag := field.Tag.Get("validate")
		if tag == "" {
			return
		}

		switch tag {
		case "required":
			if value.Kind() == reflect.String && value.String() == "" {
				errMsgs = append(errMsgs, fmt.Sprintf("%s is required", path))
			}
		case "gt=0":
			if value.Kind() == reflect.Int && value.Int() <= 0 {
				errMsgs = append(errMsgs, fmt.Sprintf("%s must be greater than 0", path))
			}
		}
	}

	// Recursively validate all fields
	var validateStruct func(v reflect.Value, path string)
	validateStruct = func(v reflect.Value, path string) {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return
		}

		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			value := v.Field(i)
			fieldPath := path
			if fieldPath != "" {
				fieldPath += "."
			}
			fieldPath += strings.ToLower(field.Name)

			if value.Kind() == reflect.Struct {
				validateStruct(value, fieldPath)
			} else {
				validateField(value, field, fieldPath)
			}
		}
	}

	validateStruct(reflect.ValueOf(c), "")

	if len(errMsgs) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errMsgs, "; "))
	}
	return nil
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

	// Validate the configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}
