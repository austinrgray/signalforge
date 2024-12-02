package main

import (
	"fmt"
	"log"
	"signalforge/normandySR1/subsystems/habitat/oxygenRecycler/config"
	"signalforge/normandySR1/subsystems/habitat/oxygenRecycler/connection"
	"signalforge/normandySR1/subsystems/habitat/oxygenRecycler/device"
	"signalforge/normandySR1/subsystems/habitat/oxygenRecycler/test"
)

func main() {
	envConfig, err := config.LoadEnvConfig()
	if err != nil {
		log.Fatalf("Failed to load environment configuration: %v", err)
	}
	tcpServerAddress := fmt.Sprintf("%s:%s", envConfig.ServerHost, envConfig.ServerPort)
	oxygenRecycler := device.InitializeDevice()
	go test.SimulateDevice(oxygenRecycler)
	go connection.StartTCPConnection(tcpServerAddress, oxygenRecycler)

	select {}
}
