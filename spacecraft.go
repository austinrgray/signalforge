package main

import (
	"fmt"
	"log"
	"sync"
)

type spacecraft struct {
	info         *spacecraftInfo
	sensorArray  *sensorArray
	remoteBridge *remoteBridge
	engine       *engine
	fuelTank     *fuelTank
	ignition     *ignition
	console      *console
	initwg       sync.WaitGroup
	opswg        sync.WaitGroup
	lock         sync.RWMutex
}

type spacecraftInfo struct {
	id            string // uuid
	name          string
	class         class
	model         string
	status        string //enum
	position      coordinate
	flightHeading vector
	mission       mission
}

type class int

const (
	Scout class = iota
	Tanker
	Freight
	Science
	Passenger
	Colony
	Mining
	Construction
	Satellite
)

func newSpacecraft() *spacecraft {
	return &spacecraft{
		info: &spacecraftInfo{
			id:            "145",
			name:          "Normandy",
			class:         Scout,
			model:         "SR7",
			status:        "initializing",
			position:      coordinate{x: 0, y: 0, z: 0},
			flightHeading: vector{xV: 0, yV: 0, zV: 0},
		},
		sensorArray: &sensorArray{
			sensorRange:   100.0,
			sensorObjects: make([]sensorObject, 16),
			safePath:      true,
		},
		remoteBridge: &remoteBridge{
			remoteHostAddr: "127.0.0.1",
			rch:            make(chan []byte, 8),
			wch:            make(chan message, 8),
			hbInt:          1,
		},
		engine: &engine{},
		fuelTank: &fuelTank{
			empty:   false,
			maxFuel: 100.0,
			remFuel: 100.0,
		},
	}
}

func (s *spacecraft) initialize() error {
	log.Printf("initializing %s...\n", s.info.name)

	s.initwg.Add(1)
	go s.sensorArray.start(s)
	s.initwg.Wait()

	s.initwg.Add(1)
	err := s.remoteBridge.initialize(s)
	if err != nil {
		return fmt.Errorf("initialization: failed to initialize connection %w", err)
	}
	s.initwg.Wait()

	s.initwg.Add(1)
	go s.remoteBridge.connect(s)
	s.initwg.Wait()

	s.initwg.Add(1)
	go s.console.start(s)
	s.initwg.Wait()

	return nil
}

type ignition struct{}

func (i ignition) start(s *spacecraft) {}
func (i ignition) stop(s *spacecraft)  {}

type sensorArray struct {
	sensorRange   float64
	sensorObjects []sensorObject
	safePath      bool
}

type sensorObject struct {
	coord           coordinate
	deltaDistance   float64
	relativeV       vector
	projectedImpact bool
}

func (sA *sensorArray) start(s *spacecraft) {

}
