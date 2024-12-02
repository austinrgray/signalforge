package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type EnvConfig struct {
	ServerHost string
	ServerPort string
}

func LoadEnvConfig() (*EnvConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	envConfig := &EnvConfig{
		ServerHost: os.Getenv("SERVER_HOST"),
		ServerPort: os.Getenv("SERVER_PORT_TCP"),
	}

	if envConfig.ServerHost == "" || envConfig.ServerPort == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return envConfig, nil
}
