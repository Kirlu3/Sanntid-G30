package slave

import "github.com/Kirlu3/Sanntid-G30/heislab/config"

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ElevatorDirection int

const (
	D_Down ElevatorDirection = -1
	D_Stop ElevatorDirection = 0
	D_Up   ElevatorDirection = 1
)

type Elevator struct {
	Floor     int
	Direction ElevatorDirection
	Requests  [config.N_FLOORS][config.N_BUTTONS]bool
	Behaviour ElevatorBehaviour
	Stuck     bool
	ID        int //ID int vs Id string ???
}

// type ExpandedElevator struct {
// 	Elevator Elevator
// 	CabCalls [N_FLOORS][2]bool // the master doesnt care about the Requests attribute of the Elevator but needs a way to store cab and hall calls
// }

type Calls struct {
	HallCalls [config.N_FLOORS][config.N_BUTTONS - 1]bool
	CabCalls  [config.N_ELEVATORS][config.N_FLOORS]bool // the master doesnt care about the Requests attribute of the Elevator but needs a way to store cab and hall calls
}

type WorldView struct {
	Elevators      [config.N_ELEVATORS]Elevator //
	OwnId          string
	HallCalls      [config.N_FLOORS][config.N_BUTTONS - 1]bool
	CabCalls       [config.N_ELEVATORS][config.N_FLOORS]bool // the master doesnt care about the Requests attribute of the Elevator but needs a way to store cab and hall calls
	AliveElevators [config.N_ELEVATORS]bool
}

type BackupCalls struct {
	Calls Calls
	Id    int
}
