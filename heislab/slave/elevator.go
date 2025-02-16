package Slave

const N_BUTTONS = 3
const N_FLOORS = 4

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
	Requests  [N_FLOORS][N_BUTTONS]bool
	Behaviour ElevatorBehaviour
	Stuck     bool
	Id        string
}

// probably better to just add id to elevator struct
// type ExpandedElevator struct {
// 	Elevator Elevator
// 	Id       string
// }

type WorldView struct {
	Elevators []Elevator
	OwnId     string
}
