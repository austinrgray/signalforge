package connection

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"signalforge/normandySR1/subsystems/habitat/oxygenRecycler/device"
	"time"
)

func StartTCPConnection(tcpServerAddress string, oxygenRecycler *device.OxygenRecycler) (net.Conn, error) {
	conn, err := net.Dial("tcp", tcpServerAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	fmt.Println("Connected to server:", tcpServerAddress)

	go sendHeartBeats(conn, oxygenRecycler)

	return conn, nil

}

func sendHeartBeats(conn net.Conn, oxygenRecycler *device.OxygenRecycler) {
	for {
		heartbeat := struct {
			DeviceID         string  `json:"device_id"`
			Temperature      float32 `json:"temperature"`
			Pressure         float32 `json:"pressure"`
			O2Output         float32 `json:"o2_output"`
			O2Concentration  float32 `json:"o2_concentration"`
			CO2Input         float32 `json:"co2_input"`
			CO2Concentration float32 `json:"co2_concentration"`
			PowerConsumption float32 `json:"power_consumption"`
			AlertLevel       string  `json:"alert_level"`
		}{
			DeviceID:         oxygenRecycler.DeviceID,
			Temperature:      oxygenRecycler.Temperature,
			Pressure:         oxygenRecycler.Pressure,
			O2Output:         oxygenRecycler.O2Output,
			O2Concentration:  oxygenRecycler.O2Concentration,
			CO2Input:         oxygenRecycler.CO2Input,
			CO2Concentration: oxygenRecycler.CO2Concentration,
			PowerConsumption: oxygenRecycler.PowerConsumption,
			AlertLevel:       oxygenRecycler.AlertLevel,
		}

		// Serialize the heartbeat payload into JSON format
		heartbeatData, err := json.Marshal(heartbeat)
		if err != nil {
			log.Printf("Failed to marshal heartbeat data: %v", err)
			return
		}

		// Send the serialized data to the TCP server
		_, err = conn.Write(heartbeatData)
		if err != nil {
			log.Printf("Failed to send heartbeat: %v", err)
			return
		}

		log.Printf("Sent heartbeat: %+v", heartbeat)

		// Wait for the heartbeat interval before sending the next one
		time.Sleep(time.Duration(oxygenRecycler.HeartbeatInterval) * time.Second)
	}
}
