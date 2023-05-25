package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Config struct {
	ActiveSymbol string `json:"active_symbol"`
}

func loadConfig() (Config, error) {
	configPath := path.Join(getHomeDir(), getConfigFileName())
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	unmarshalErr := json.Unmarshal(configFile, &config)
	if unmarshalErr != nil {
		return Config{}, unmarshalErr
	}

	return config, nil
}

func writeConfig(config Config) error {
	newConfigBytes, marshalErr := json.Marshal(config)
	if marshalErr != nil {
		return marshalErr
	}

	configPath := path.Join(getHomeDir(), getConfigFileName())
	createConfigErr := os.WriteFile(configPath, newConfigBytes, 0644)
	if createConfigErr != nil {
		return createConfigErr
	}

	return nil
}

func getActiveSymbol() (string, error) {
	config, err := loadConfig()
	if err != nil {
		return "", err
	}

	return config.ActiveSymbol, nil
}

func getToken() (string, error) {
	activeSymbol, activeSymbolErr := getActiveSymbol()
	if activeSymbolErr != nil {
		return "", activeSymbolErr
	}
	tokenFileName := fmt.Sprintf("%s.token", activeSymbol)
	tokenPath := path.Join(getHomeDir(), tokenFileName)
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getHomeDir() string {
	// Make a new directory at a known good path for a token + config
	dirname, dirErr := os.UserHomeDir()
	if dirErr != nil {
		panic(dirErr)
	}

	return path.Join(dirname, ".spacetraders")
}

func getConfigFileName() string {
	return "config.json"
}

func isInitialized() bool {
	token, err := getToken()
	if err != nil {
		return false
	}

	return len(token) > 0
}
