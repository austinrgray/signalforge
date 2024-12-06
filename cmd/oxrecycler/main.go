package main

import (
	"flag"
	"log"
	"signalforge/pkg/oxrecycler"
)

func main() {
	/*err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	tcpServerAddress := os.Getenv("SERVER_HOST_TCP") + ":" + os.Getenv("SERVER_PORT_TCP")
	*/
	deviceConfig := flag.String("device", "primary", "Key of the device to load from the configuration (e.g., PrimaryRecycler, SecondaryRecycler)")
	flag.Parse()

	tcpServerAddress := "localhost:9000"

	device, err := oxrecycler.LoadDeviceFromConfig(*deviceConfig)
	if err != nil {
		log.Fatalf("Error loading device from config: %v", err)
	}

	device.Start(tcpServerAddress)

	select {}
}
