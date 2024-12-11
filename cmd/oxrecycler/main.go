package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"signalforge/pkg/oxrecycler"
)

func main() {
	presetOption := flag.String("device", "1", "Key of the preset device to use for testing (e.g., 1, 2, 3")
	flag.Parse()

	config, err := oxrecycler.LoadConfigs()
	if err != nil {
		log.Fatalf("Error loading device from config: %v", err)
	}

	var device oxrecycler.Device
	switch *presetOption {
	case "1":
		device = oxrecycler.Device{
			Config:            *config,
			DeviceID:          "oxr1234",
			Temperature:       35.3,
			Mode:              "NORMAL",
			LastMaintenance:   time.Time{},
			HeartbeatInterval: 1 * time.Second,
			Errors:            make([]oxrecycler.Error, 0),
		}
	case "2":
		device = oxrecycler.Device{
			Config:            *config,
			DeviceID:          "oxr5678",
			Temperature:       16.9,
			Mode:              "AUTO",
			LastMaintenance:   time.Time{},
			HeartbeatInterval: 2 * time.Second,
			Errors: []oxrecycler.Error{
				{Code: "E1001", Description: "Warning: device needs calibrated", AlertLevel: "Low", Timestamp: time.Now()}},
		}
	case "3":
		device = oxrecycler.Device{
			Config:            *config,
			DeviceID:          "oxr9101",
			Temperature:       0.0,
			Mode:              "MAINTENANCE",
			LastMaintenance:   time.Now(),
			HeartbeatInterval: 3 * time.Second,
			Errors: []oxrecycler.Error{
				{Code: "E1001", Description: "Warning: device needs calibrated", AlertLevel: "Low", Timestamp: time.Now()},
				{Code: "E7001", Description: "Faulty Intake", AlertLevel: "CRITICAL", Timestamp: time.Now()}},
		}
	default:
		fmt.Printf("Could not read from config")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("Starting device: %s", device.DeviceID)
		device.Start()
	}()

	wg.Wait()
}
