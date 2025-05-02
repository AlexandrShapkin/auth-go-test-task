package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

// Загружает файл конфигурации.
// 
// Для подробной спецификации параметров см. github.com/spf13/viper (SetConfigName, SetConfigType, AddConfigPath)
//
// !! Не обрабатывает переменные окружения, хотя и может написать предупреждение указывающее на возможность работы с ними !!
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