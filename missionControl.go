package main

type mission struct {
	id             string //uuid
	status         missionStatus
	manifests      []manifest
	waypoints      []waypoint
	directives     []directive
	totalTimeSteps int
}

type missionStatus struct {
	curManifest      manifest
	curWaypoint      waypoint
	curDirective     directive
	deltaTimesteps   int //next waypoint or total remaining?
	enoughProvisions bool
	enoughFuel       bool
}

type manifest struct {
	id           string //uuid
	recWaypoint  waypoint
	delWapypoint waypoint
	souls        uint
	provisions   uint               // units in earthdays/soul
	cargo        map[string]float32 //key=cargoitem_units / value=amount in unit
}

type waypoint struct {
	id             string   //uuid
	name           string   //if applicable
	missionCommand []string //enum(s) if applicable
	coord          coordinate
	deltaDistance  float64
	deltaTimesteps int
}

type directive struct {
	timeStep        int
	directedheading vector
	stepInstruction instruction //if applicable
	estStepFuelCost float32
	estRemTimeSteps int
	waypoint        waypoint
}

// ---------------- Traversal ----------------

type coordinate struct {
	x float64
	y float64
	z float64
}

type vector struct {
	xV float64
	yV float64
	zV float64
}
