package test

import (
	"log"
	"signalforge/normandySR1/subsystems/habitat/oxygenRecycler/device"
	"time"
)

func SimulateDevice(oxygenRecycler *device.OxygenRecycler) {
	for {
		oxygenRecycler.MutateDevice()
		log.Printf("Simulated device mutation: %+v", *oxygenRecycler)
		time.Sleep(1 * time.Second)
	}
}
