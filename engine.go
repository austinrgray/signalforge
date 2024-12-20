package main

type engine struct{}

// units of fuel should relate to cost of heading changes for easu calculation of pathing constraints
type fuelTank struct {
	empty   bool
	maxFuel float32
	remFuel float32
}
