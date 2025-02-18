package slave

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

// type ExpandedElevator struct {
// 	Elevator Elevator
// 	CabCalls [N_FLOORS][2]bool // the master doesnt care about the Requests attribute of the Elevator but needs a way to store cab and hall calls
// }

type Calls struct {
	HallCalls [N_FLOORS]bool
	CabCalls  [10][N_FLOORS][2]bool // the master doesnt care about the Requests attribute of the Elevator but needs a way to store cab and hall calls
}

type WorldView struct {
	Elevators      [10]Elevator //
	OwnId          string
	HallCalls      [N_FLOORS]bool
	CabCalls       [10][N_FLOORS][2]bool // the master doesnt care about the Requests attribute of the Elevator but needs a way to store cab and hall calls
	AliveElevators [10]bool
}
