package main

import (
	"signalforge/pkg/oxrecycler"
)

func main() {
	/*err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	tcpServerAddress := os.Getenv("SERVER_HOST_TCP") + ":" + os.Getenv("SERVER_PORT_TCP")
	*/
	tcpServerAddress := "localhost:9000"
	device := oxrecycler.DefaultDevice()

	device.Start(tcpServerAddress)

	select {}
}
