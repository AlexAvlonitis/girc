// Package config loads the client configuration from a YAML file.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	defaultPort    = 6667
	defaultTLSPort = 6697
)

// Config holds the settings needed to connect to an IRC server.
type Config struct {
	Server   string `yaml:"server"`
	Port     int    `yaml:"port"`
	Nick     string `yaml:"nick"`
	User     string `yaml:"user"`
	RealName string `yaml:"realName"`
	TLS      bool   `yaml:"ssl"`
	// InsecureSkipVerify disables TLS certificate verification. Only use it
	// for servers with self-signed certificates that you trust.
	InsecureSkipVerify bool `yaml:"insecureSkipVerify"`
}

// Load reads and validates the configuration file at path.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", path, err)
	}

	cfg.setDefaults()
	if err := cfg.validate(); err != nil {
		return Config{}, fmt.Errorf("invalid config %s: %w", path, err)
	}
	return cfg, nil
}

// Addr returns the server address in host:port form.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Server, strconv.Itoa(c.Port))
}

func (c *Config) setDefaults() {
	if c.Port == 0 {
		c.Port = defaultPort
		if c.TLS {
			c.Port = defaultTLSPort
		}
	}
	if c.User == "" {
		c.User = c.Nick
	}
	if c.RealName == "" {
		c.RealName = c.Nick
	}
}

func (c Config) validate() error {
	var errs []error
	if c.Server == "" {
		errs = append(errs, errors.New("server is required"))
	}
	if c.Nick == "" {
		errs = append(errs, errors.New("nick is required"))
	}
	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("port %d out of range", c.Port))
	}
	return errors.Join(errs...)
}
