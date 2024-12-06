package oxrecycler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type Device struct {
	ID                  string        `json:"ID"`
	SerialNumber        string        `json:"SerialNumber"`
	Status              string        `json:"Status"`
	ConnectionStatus    string        `json:"ConnectionStatus"`
	Mode                string        `json:"Mode"`
	Temperature         float32       `json:"Temperature"`
	Pressure            float32       `json:"Pressure"`
	O2Output            float32       `json:"O2Output"`
	O2Concentration     float32       `json:"O2Concentration"`
	CO2Input            float32       `json:"CO2Input"`
	CO2Concentration    float32       `json:"CO2Concentration"`
	PowerConsumption    float32       `json:"PowerConsumption"`
	AlertLevel          string        `json:"AlertLevel"`
	ErrorCodes          []string      `json:"ErrorCodes"`
	ErrorMessages       []string      `json:"ErrorMessages"`
	LastMaintenance     *time.Time    `json:"LastMaintenance"`
	LastCommTime        *time.Time    `json:"LastCommTime"`
	HeartbeatInterval   time.Duration `json:"HeartbeatInterval"`
	connectionContext   context.Context
	connectionCancel    context.CancelFunc
	simulationContext   context.Context
	simulationCancel    context.CancelFunc
	tcpServerConnection net.Conn
}

func LoadDeviceFromConfig(deviceType string) (*Device, error) {
	file, err := os.Open("pkg/oxrecycler/config.json")
	if err != nil {
		return nil, fmt.Errorf("could not open config.json: %w", err)
	}
	defer file.Close()

	var config map[string]struct {
		DefaultValues Device `json:"default_values"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("could not parse JSON from file %s: %w", file.Name(), err)
	}

	var deviceKey string
	switch deviceType {
	case "primary":
		deviceKey = "o2_recycler_primary"
	case "secondary":
		deviceKey = "o2_recycler_secondary"
	default:
		return nil, fmt.Errorf("invalid device type: %s", deviceType)
	}

	deviceConfig, exists := config[deviceKey]
	if !exists {
		return nil, fmt.Errorf("device type %s not found in configuration", deviceType)
	}

	device := deviceConfig.DefaultValues
	device.HeartbeatInterval *= time.Second

	if device.LastMaintenance == nil {
		device.LastMaintenance = &time.Time{} // Default value
	}
	if device.LastCommTime == nil {
		device.LastCommTime = &time.Time{} // Default value
	}

	return &device, nil
}

func (d *Device) Start(tcpServerAddress string) {
	d.connectionContext, d.connectionCancel = context.WithCancel(context.Background())
	d.simulationContext, d.simulationCancel = context.WithCancel(context.Background())
	go startSimulation(d)
	go d.NewTCPServerConnection(tcpServerAddress)
}

func (d *Device) NewTCPServerConnection(tcpServerAddress string) {
	for {
		select {
		case <-d.connectionContext.Done():
			if d.tcpServerConnection != nil {
				d.tcpServerConnection.Close()
			}
			fmt.Println("Connection stopped for device:", d.ID)
			return
		default:
			if d.tcpServerConnection == nil {
				conn, err := net.Dial("tcp", tcpServerAddress)
				if err != nil {
					fmt.Println("Failed to connect:", err)
					time.Sleep(5 * time.Second)
					continue
				}
				d.tcpServerConnection = conn
				fmt.Println("Device connected to server:", d.ID)

				go d.startHeartbeat()
				go d.listenForServerMessages()
			}
		}
	}
}

func (d *Device) listenForServerMessages() {
	for {
		buffer := make([]byte, 1024)

		n, err := d.tcpServerConnection.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from connection:", err)
			} else {
				fmt.Println("Server closed connection for device:", d.ID)
				d.tcpServerConnection.Close()
				d.tcpServerConnection = nil
			}
			time.Sleep(2 * time.Second)
			continue
		}

		message := string(buffer[:n])
		fmt.Println("Received message for device:", d.ID, "Message:", message)
	}
}

func (d *Device) startHeartbeat() {
	retries := 0
	for {
		if d.tcpServerConnection == nil {
			fmt.Println("No active TCP connection for device:", d.ID)
			return
		}

		_, err := d.tcpServerConnection.Write([]byte(d.heartbeatMessage()))
		if err != nil {
			fmt.Println("Error sending heartbeat:", err)

			retries++

			if retries >= 15 {
				fmt.Println("Retry limit reached, connection might be unstable.")
				d.tcpServerConnection.Close()
				d.tcpServerConnection = nil
				return
			}
			continue
		}
		time.Sleep(d.HeartbeatInterval)
	}
}

func (d *Device) Stop() {
	d.DisconnectTCPServer()
	d.StopSimulation()
}

func (d *Device) DisconnectTCPServer() {
	d.connectionCancel()
}

func (d *Device) StopSimulation() {
	d.simulationCancel()
}

func (d *Device) heartbeatMessage() string {
	hMsg := struct {
		ID               string   `json:"id"`
		SerialNumber     string   `json:"serial_number"`
		Status           string   `json:"status"`
		MessageType      string   `json:"messageType"`
		ConnectionStatus string   `json:"connection_status"`
		Mode             string   `json:"mode"`
		Temperature      float32  `json:"temperature"`
		Pressure         float32  `json:"pressure"`
		O2Output         float32  `json:"o2_output"`
		O2Concentration  float32  `json:"o2_concentration"`
		CO2Input         float32  `json:"co2_input"`
		CO2Concentration float32  `json:"co2_concentration"`
		PowerConsumption float32  `json:"power_consumption"`
		AlertLevel       string   `json:"alert_level"`
		ErrorCodes       []string `json:"error_codes"`
		ErrorMessages    []string `json:"error_messages"`
		LastMaintenance  string   `json:"last_maintenance"`
		LastCommTime     string   `json:"last_comm_time"`
	}{
		ID:               d.ID,
		SerialNumber:     d.SerialNumber,
		Status:           d.Status,
		MessageType:      "heartbeat",
		ConnectionStatus: d.ConnectionStatus,
		Mode:             d.Mode,
		Temperature:      d.Temperature,
		Pressure:         d.Pressure,
		O2Output:         d.O2Output,
		O2Concentration:  d.O2Concentration,
		CO2Input:         d.CO2Input,
		CO2Concentration: d.CO2Concentration,
		PowerConsumption: d.PowerConsumption,
		AlertLevel:       d.AlertLevel,
		ErrorCodes:       d.ErrorCodes,
		ErrorMessages:    d.ErrorMessages,
		LastMaintenance:  d.LastMaintenance.Format("2006-01-02 15:04:05"),
		LastCommTime:     d.LastCommTime.Format("2006-01-02 15:04:05"),
	}

	jsonData, err := json.MarshalIndent(hMsg, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return ""
	}
	return string(jsonData)
}

func (d *Device) MutateDevice() {
	d.Temperature += 0.05
	d.Pressure += 0.02
	d.O2Concentration += 0.01
	d.CO2Concentration -= 0.005
	d.PowerConsumption += 0.03
	d.O2Output -= 0.02
	d.CO2Input += 0.01
}
