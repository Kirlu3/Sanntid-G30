package elevator

import "Driver-go/elevio"

type Behaviour int

const (
	Idle Behaviour = iota
	DoorOpen
	Moving	
)


type Elevator struct {
	Floor		int 						
	Dirn		elevio.MotorDirection
	Requests 	[elevio._numFloors][_numFloors]Requests
}