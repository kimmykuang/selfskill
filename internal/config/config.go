package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  int            `yaml:"version"`
	Registry RegistryConfig `yaml:"registry"`
}

type RegistryConfig struct {
	Default string `yaml:"default"`
}

func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Registry: RegistryConfig{
			Default: "",
		},
	}
}

// Load reads the global config from ~/ss/ss.yaml.
// If the file doesn't exist, returns a default config.
func Load() (*Config, error) {
	path := ConfigFile()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to ~/ss/ss.yaml.
func (c *Config) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile(), data, 0644)
}
