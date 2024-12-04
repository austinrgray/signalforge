package oxrecycler

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

type Device struct {
	ID                  string
	SerialNumber        string
	Status              string
	ConnectionStatus    string
	Mode                string
	Temperature         float32
	Pressure            float32
	O2Output            float32
	O2Concentration     float32
	CO2Input            float32
	CO2Concentration    float32
	PowerConsumption    float32
	AlertLevel          string
	ErrorCodes          []string
	ErrorMessages       []string
	LastMaintenance     time.Time
	LastCommTime        time.Time
	Runtime             time.Duration
	HeartbeatInterval   time.Duration
	connectionContext   context.Context
	connectionCancel    context.CancelFunc
	simulationContext   context.Context
	simulationCancel    context.CancelFunc
	tcpServerConnection net.Conn
}

func DefaultDevice() *Device {
	return &Device{
		ID:                "O2-Habitat-Primary",
		SerialNumber:      "O2R-SN4567",
		Status:            "Offline",
		ConnectionStatus:  "No Connection",
		Mode:              "Normal",
		Temperature:       22.5,
		Pressure:          1.2,
		O2Output:          10.0,
		O2Concentration:   21.0,
		CO2Input:          0.0,
		CO2Concentration:  0.04,
		PowerConsumption:  15.0,
		AlertLevel:        "Normal",
		ErrorCodes:        []string{},
		ErrorMessages:     []string{},
		LastMaintenance:   time.Now(),
		LastCommTime:      time.Now(),
		Runtime:           0,
		HeartbeatInterval: 5 * time.Second,
	}
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

		// Process the received message
		message := string(buffer[:n])
		fmt.Println("Received message for device:", d.ID, "Message:", message)
	}
}

func (d *Device) startHeartbeat() {
	retries := 0
	for {
		if d.tcpServerConnection == nil {
			// Skip heartbeat if the connection is not available
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
	return fmt.Sprintf(`
		----------------------------
		Heartbeat for Device: %s
		----------------------------
		Serial Number:      %s
		Status:             %s
		Connection Status:  %s
		Mode:               %s
		Temperature:        %.2f °C
		Pressure:           %.2f hPa
		O2 Output:          %.2f L/min
		O2 Concentration:   %.2f %%
		CO2 Input:          %.2f L/min
		CO2 Concentration:  %.2f %%
		Power Consumption:  %.2f W
		Alert Level:        %s
		Error Codes:        %v
		Error Messages:     %v
		Last Maintenance:   %s
		Last Communication: %s
		Runtime:            %s
		----------------------------
		`,
		d.ID,
		d.SerialNumber,
		d.Status,
		d.ConnectionStatus,
		d.Mode,
		d.Temperature,
		d.Pressure,
		d.O2Output,
		d.O2Concentration,
		d.CO2Input,
		d.CO2Concentration,
		d.PowerConsumption,
		d.AlertLevel,
		d.ErrorCodes,
		d.ErrorMessages,
		d.LastMaintenance.Format("2006-01-02 15:04:05"),
		d.LastCommTime.Format("2006-01-02 15:04:05"),
		d.Runtime.String(),
	)
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
