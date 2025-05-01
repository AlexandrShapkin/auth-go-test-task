package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

func LoadConfig(configName string, configType string, path string) (*Config, error) {
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AddConfigPath(path)

	err := viper.ReadInConfig()
	if err != nil {
		slog.Warn("Failed to load configuration. Envs only", "error", err)
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}