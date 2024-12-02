package device

import (
	//"fmt"
	"log"
	"time"
)

type OxygenRecycler struct {
	DeviceID          string
	SerialNumber      string
	Status            string
	ConnectionStatus  string
	Mode              string
	Temperature       float32
	Pressure          float32
	O2Output          float32
	O2Concentration   float32
	CO2Input          float32
	CO2Concentration  float32
	PowerConsumption  float32
	AlertLevel        string
	ErrorCodes        []string
	ErrorMessages     []string
	LastMaintenance   time.Time
	LastCommTime      time.Time
	Runtime           time.Duration
	HeartbeatInterval time.Duration
}

// NewOxygenRecycler creates a new OxygenRecycler device with default values.
func InitializeDevice() *OxygenRecycler {
	return &OxygenRecycler{
		DeviceID:          "O2-Habitat-Primary",
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
		HeartbeatInterval: 5,
	}
}

func (d *OxygenRecycler) DisplayDeviceInfo() {
	log.Printf("Device ID: %s", d.DeviceID)
	log.Printf("Serial Number: %s", d.SerialNumber)
	log.Printf("Status: %s", d.Status)
	log.Printf("Connection Status: %s", d.ConnectionStatus)
	log.Printf("Mode: %s", d.Mode)
	log.Printf("Temperature: %.2f°C", d.Temperature)
	log.Printf("Pressure: %.2f bar", d.Pressure)
	log.Printf("O2 Output: %.2f L/min", d.O2Output)
	log.Printf("O2 Concentration: %.2f%%", d.O2Concentration)
	log.Printf("CO2 Input: %.2f L/min", d.CO2Input)
	log.Printf("CO2 Concentration: %.2f%%", d.CO2Concentration)
	log.Printf("Power Consumption: %.2f W", d.PowerConsumption)
	log.Printf("Alert Level: %s", d.AlertLevel)
	log.Printf("Last Maintenance: %s", d.LastMaintenance.Format(time.RFC3339))
	log.Printf("Last Communication Time: %s", d.LastCommTime.Format(time.RFC3339))
	log.Printf("Runtime: %d seconds", d.Runtime)
	log.Printf("Heartbeat Interval: %d seconds", d.HeartbeatInterval)
}

func (d *OxygenRecycler) UpdateHeartbeat() {
	log.Printf("Sending heartbeat with current device data...")
	d.DisplayDeviceInfo() // Display current device data in the heartbeat
}

func (d *OxygenRecycler) MutateDevice() {
	// Mutate device parameters based on the coefficients (simplified here for the example)
	d.Temperature += 0.05 // Simulating a fluctuation
	d.Pressure += 0.02
	d.O2Concentration += 0.01
	d.CO2Concentration -= 0.005
	d.PowerConsumption += 0.03
	d.O2Output -= 0.02
	d.CO2Input += 0.01
}
