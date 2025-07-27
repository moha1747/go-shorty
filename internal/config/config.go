package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

// Config holds the global application configuration
type Config struct {
	DNS      DNSConfig      `mapstructure:"dns"`
	Redirect RedirectConfig `mapstructure:"redirect"`
}

// DNSConfig holds DNS server configuration
type DNSConfig struct {
	Port        int    `mapstructure:"port"`
	UpstreamDNS string `mapstructure:"upstream_dns"`
	LocalIP     string `mapstructure:"local_ip"`
}

// RedirectConfig holds HTTP redirect server configuration
type RedirectConfig struct {
	Port      int               `mapstructure:"port"`
	Address   string            `mapstructure:"address"`
	Shortcuts map[string]string `mapstructure:"shortcuts"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DNS: DNSConfig{
			Port:        53,
			UpstreamDNS: "1.1.1.1:53",
			LocalIP:     "127.0.0.1",
		},
		Redirect: RedirectConfig{
			Port:    80,
			Address: "127.0.0.1",
			Shortcuts: map[string]string{
				"go": "https://go.dev",
				"gh": "https://github.com",
				"so": "https://stackoverflow.com",
			},
		},
	}
}

// LoadConfig loads the configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// Set up viper
	v := viper.New()
	v.SetConfigName("config") // name of config file (without extension)
	v.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name

	if configPath != "" {
		v.AddConfigPath(configPath)
	}

	// Add default paths to look for the config file
	v.AddConfigPath(".")                // look for config in the working directory
	v.AddConfigPath("./config")         // look for config in ./config/ directory
	v.AddConfigPath("$HOME/.go-shorty") // look in home directory

	// Set up environment variables
	v.SetEnvPrefix("GOSHORTY")
	v.AutomaticEnv() // read in environment variables that match

	// Set default values
	setDefaultsFromConfig(v, config)

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		log.Println("No config file found, using defaults")
	}

	// Unmarshal the config
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return config, nil
}

// setDefaultsFromConfig sets the default values in Viper from the default config
func setDefaultsFromConfig(v *viper.Viper, config *Config) {
	// DNS defaults
	v.SetDefault("dns.port", config.DNS.Port)
	v.SetDefault("dns.upstream_dns", config.DNS.UpstreamDNS)
	v.SetDefault("dns.local_ip", config.DNS.LocalIP)

	// Redirect defaults
	v.SetDefault("redirect.port", config.Redirect.Port)
	v.SetDefault("redirect.address", config.Redirect.Address)

	// Set shortcuts
	for key, value := range config.Redirect.Shortcuts {
		v.SetDefault(fmt.Sprintf("redirect.shortcuts.%s", key), value)
	}
}

// WriteDefaultConfig writes the default configuration to a file
func WriteDefaultConfig(filePath string) error {
	v := viper.New()
	config := DefaultConfig()

	// Set up the defaults
	setDefaultsFromConfig(v, config)

	// Set the config file
	v.SetConfigFile(filePath)

	// Save the config
	return v.WriteConfig()
}
