package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const ConfigPath = "~/.95cli/config.json"

// DefaultAPIURL is the production API endpoint
const DefaultAPIURL = "https://api.95ninefive.dev"

// LocalAPIURL is the development API endpoint
const LocalAPIURL = "http://localhost:8080"

type Config struct {
	APIUrl       string `json:"api_url" mapstructure:"api_url"`
	AccessToken  string `json:"access_token" mapstructure:"access_token"`
	RefreshToken string `json:"refresh_token" mapstructure:"refresh_token"`
	UserId       int    `json:"user_id" mapstructure:"user_id"`
	Username     string `json:"username" mapstructure:"username"`
}

// GetAPIURL returns the API URL
func (cfg *Config) GetAPIURL() string {
	if os.Getenv("DEV_MODE") == "true" {
		return LocalAPIURL
	}
	return DefaultAPIURL
}

type ProjectConfig struct {
	RunCommand string `json:"runCommand"`
	Language   string `json:"language"`
}

func Init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	configDir := filepath.Join(home, ".95cli")
	err = os.MkdirAll(configDir, 0700)
	if err != nil {
		panic(err)
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("json")
}

func Load() (*Config, error) {
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return &Config{}, nil
		}
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg *Config) Save() error {
	// marshal struct to map for Viper
	viper.Set("api_url", cfg.APIUrl)
	viper.Set("access_token", cfg.AccessToken)
	viper.Set("refresh_token", cfg.RefreshToken)
	viper.Set("user_id", cfg.UserId)
	viper.Set("username", cfg.Username)

	// creates if doesn't exist
	err := viper.SafeWriteConfig()
	if err != nil {
		// if file exists, we overwrite
		return viper.WriteConfig()
	}
	return nil
}

func Clear() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// deletes the whole .95cli dir
	// per docs: If the path does not exist, RemoveAll returns nil (no error)
	if err := os.RemoveAll(filepath.Join(homeDir, ".95cli")); err != nil {
		return err
	}
	return nil
}

// LoadProjectConfig reads .95cli-project.json from current directory
func LoadProjectConfig() (*ProjectConfig, error) {
	currDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	v := viper.New()
	v.AddConfigPath(currDir)
	v.SetConfigFile("config")
	v.SetConfigType("json")

	err = v.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return &ProjectConfig{}, nil
		}
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var cfg ProjectConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal project config: %w", err)
	}
	return &cfg, nil
}

// SaveProjectConfig writes runCommand and language to config.json in current directory
func SaveProjectConfig(runCommand string, language string) error {
	currDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	v := viper.New()
	v.AddConfigPath(currDir)
	v.SetConfigFile("config")
	v.SetConfigType("json")
	v.Set("runCommand", runCommand)
	v.Set("language", language)

	// safeWriteConfig does not seem to be safe!
	err = v.SafeWriteConfig()
	if err != nil {
		return v.WriteConfig()
	}
	return nil
}

// TODO: Add other supported languages (Zig, Clojure, Ruby...)
// DetectLanguage detects programming language from run command
func DetectLanguage(runCommand string) string {
	// Simple detection based on command prefix
	switch {
	case len(runCommand) >= 2 && runCommand[:2] == "go":
		return "GO"
	case len(runCommand) >= 6 && runCommand[:6] == "python":
		return "PYTHON"
	case len(runCommand) >= 4 && runCommand[:4] == "java":
		return "JAVA"
	case len(runCommand) >= 4 && runCommand[:4] == "rust", len(runCommand) >= 5 && runCommand[:5] == "cargo":
		return "RUST"
	case len(runCommand) >= 4 && runCommand[:4] == "node":
		return "JAVASCRIPT"
	case len(runCommand) >= 3 && runCommand[:3] == "g++", len(runCommand) >= 5 && runCommand[:5] == "clang":
		return "CPP"
	case len(runCommand) >= 3 && runCommand[:3] == "gcc":
		return "C"
	default:
		return "UNKNOWN"
	}
}
