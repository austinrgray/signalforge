package oxrecycler

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

const MaxConnectionAttempts = 5
const ConnectionAttemptDelay = 3

type Config struct {
	TCPHost string
	TCPPort string
	Device  *Device
	Devices []*Device `json:"devices"`
}

func LoadConfigs(jsonConfigPath string, presetID string) (*Config, error) {
	var config Config

	err := config.LoadEnvConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	err = config.LoadJSONConfig(jsonConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	err = config.LoadDevicePreset(presetID)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}
	return &config, nil
}

func (c *Config) LoadEnvConfig() error {
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("Error loading .env file: %v", err)
	}
	c.TCPHost = os.Getenv("SERVER_HOST_TCP")
	c.TCPPort = os.Getenv("SERVER_PORT_TCP")

	if c.TCPHost == "" || c.TCPPort == "" {
		return fmt.Errorf("missing required environment variables: SERVER_HOST_TCP or SERVER_PORT_TCP")
	}
	return nil
}

func (c *Config) LoadJSONConfig(jsonConfigPath string) error {
	data, err := os.ReadFile(jsonConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	err = json.Unmarshal(data, &c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}
	return nil
}

func (c *Config) LoadDevicePreset(id string) error {
	for _, d := range c.Devices {
		if d.ID == id {
			c.Device = d
			return nil
		}
	}
	return fmt.Errorf("device with ID %s not found", id)
}
