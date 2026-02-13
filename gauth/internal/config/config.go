// Package config handles YAML config persistence.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all persistent state.
type Config struct {
	// Device registration
	AndroidID     string `yaml:"android_id"`
	SecurityToken string `yaml:"security_token"`

	// Account
	Email       string `yaml:"email"`
	MasterToken string `yaml:"master_token"`

	// Device profile
	Device DeviceConfig `yaml:"device"`

	// Server
	ServerPort int `yaml:"server_port"`
}

// DeviceConfig holds the spoofed device identity.
type DeviceConfig struct {
	Model        string `yaml:"model"`
	Brand        string `yaml:"brand"`
	Manufacturer string `yaml:"manufacturer"`
	Device       string `yaml:"device"`
	Product      string `yaml:"product"`
	Hardware     string `yaml:"hardware"`
	Fingerprint  string `yaml:"fingerprint"`
	Bootloader   string `yaml:"bootloader"`
	BuildID      string `yaml:"build_id"`
	SDKVersion   int    `yaml:"sdk_version"`
	BuildTime    int64  `yaml:"build_time"`
}

// DefaultConfig returns config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		ServerPort: 8080,
		Device: DeviceConfig{
			Model:        "Pixel 7",
			Brand:        "google",
			Manufacturer: "Google",
			Device:       "panther",
			Product:      "panther",
			Hardware:     "tensor",
			Fingerprint:  "google/panther/panther:13/TQ3A.230901.001/10750268:user/release-keys",
			Bootloader:   "slider-1.2-9971768",
			BuildID:      "TQ3A.230901.001",
			SDKVersion:   33,
			BuildTime:    1693440000,
		},
	}
}

const configFileName = "gauth_config.yaml"

// ConfigPath returns the path to the config file (same directory as executable).
func ConfigPath() string {
	exe, err := os.Executable()
	if err != nil {
		return configFileName
	}
	return filepath.Join(filepath.Dir(exe), configFileName)
}

// Load reads config from disk. Returns default config if file doesn't exist.
func Load() *Config {
	return LoadFrom(configFileName)
}

// LoadFrom reads config from a specific path.
func LoadFrom(path string) *Config {
	cfg := DefaultConfig()
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}
	_ = yaml.Unmarshal(data, cfg)
	return cfg
}

// Save writes config to disk.
func (c *Config) Save() error {
	return c.SaveTo(configFileName)
}

// SaveTo writes config to a specific path.
func (c *Config) SaveTo(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

// HasRegistration returns true if device check-in has been done.
func (c *Config) HasRegistration() bool {
	return c.AndroidID != "" && c.SecurityToken != ""
}

// HasMasterToken returns true if login has been completed.
func (c *Config) HasMasterToken() bool {
	return c.MasterToken != "" && c.Email != ""
}

// UserAgent returns the Android WebView user agent string.
func (c *Config) UserAgent() string {
	return fmt.Sprintf(
		"Mozilla/5.0 (Linux; Android %d; %s Build/%s; wv) "+
			"AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 "+
			"Chrome/120.0.6099.230 Mobile Safari/537.36 MinuteMaid",
		c.Device.SDKVersion, c.Device.Model, c.Device.BuildID,
	)
}

// AuthUserAgent returns the user agent for /auth endpoint.
func (c *Config) AuthUserAgent() string {
	return fmt.Sprintf("GoogleAuth/1.4 (%s %s); gzip", c.Device.Device, c.Device.BuildID)
}
