package main

import (
	"flag"
	"log"
	"sync"

	"signalforge/pkg/oxrecycler"
)

func main() {
	jsonConfigPath := "pkg/oxrecycler/config.json"
	presetID := flag.String("device", "device-uuid-1234", "Key of the device to load from the configuration (e.g., device-uuid-1234, device-uuid-5678, device-uuid-9101)")
	flag.Parse()

	config, err := oxrecycler.LoadConfigs(jsonConfigPath, *presetID)
	if err != nil {
		log.Fatalf("Error loading device from config: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := config.Device.InitializeConnection("localhost:3000")
		if err != nil {
			log.Printf("Error initializing connection: %v", err)
		}
		go config.Device.Start()
	}()

	wg.Wait()
}
