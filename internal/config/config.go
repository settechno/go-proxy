package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type AppConfig struct {
	Port               int    `json:"port"`
	UserFile           string `json:"user_file"`
	UserReloadDuration int    `json:"user_reload_duration"`
	UseAuth            bool   `json:"use_auth"`
}

func GetAppConfigFile() string {
	// Проверяем переменную окружения
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}

	// Используем конфиг по умолчанию
	return "config.json"
}

func LoadAppConfig(configFile string) (*AppConfig, error) {
	log.Printf("Loading config from: %s\n", configFile)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("cannot read appConfig file: %v", err)
	}

	var appConfig AppConfig
	if err := json.Unmarshal(data, &appConfig); err != nil {
		return nil, fmt.Errorf("invalid appConfig format: %v", err)
	}

	// Проверяем обязательные поля
	if appConfig.Port == 0 {
		return nil, fmt.Errorf("port is required")
	}
	if appConfig.UserFile == "" {
		return nil, fmt.Errorf("users file is required")
	}
	if appConfig.UserReloadDuration == 0 {
		appConfig.UserReloadDuration = 2
	}

	return &appConfig, nil
}
