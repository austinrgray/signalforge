package oxrecycler

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Config struct {
	TCPServerHost          string   `json:"-"`
	TCPServerPort          string   `json:"-"`
	MaxConnectionAttempts  int      `json:"max_connection_attempts"`
	MaxRetriesOnError      int      `json:"max_retries_on_error"`
	MaxMessageSize         int      `json:"max_message_size"`
	ConnectionAttemptDelay Duration `json:"connection_attempt_delay"`
	ConnectionLockout      Duration `json:"connection_lockout_duration"`
	HandshakeTimeout       Duration `json:"handshake_timeout"`
	ReadTimeout            Duration `json:"read_timeout"`
	WriteTimeout           Duration `json:"write_timeout"`
}

type Duration time.Duration

func (d *Duration) UnmarshalJSON(b []byte) error {
	str := string(b)
	str = str[1 : len(str)-1]
	duration, err := time.ParseDuration(str)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s", str)
	}
	*d = Duration(duration)
	return nil
}

func LoadConfigs() (*Config, error) {
	//Load and Parse config.json
	configJSON, err := os.Open("pkg/oxrecycler/config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	defer configJSON.Close()

	var config Config
	decoder := json.NewDecoder(configJSON)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config.json file: %v", err)
	}

	//Load .env variables
	if host := os.Getenv("SERVER_HOST_TCP"); host != "" {
		config.TCPServerHost = host
	} else {
		config.TCPServerHost = "localhost"
	}
	if port := os.Getenv("SERVER_PORT_TCP"); port != "" {
		config.TCPServerPort = port
	} else {
		config.TCPServerHost = ":3000"
	}
	return &config, nil
}
