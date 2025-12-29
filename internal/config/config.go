package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const ConfigPath = "~/.95cli/config.json"

type Config struct {
	APIUrl       string `json:"api_url" mapstructure:"api_url"`
	AccessToken  string `json:"access_token" mapstructure:"access_token"`
	RefreshToken string `json:"refresh_token" mapstructure:"refresh_token"`
	UserId       int    `json:"user_id" mapstructure:"user_id"`
	Username     string `json:"username" mapstructure:"username"`
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
