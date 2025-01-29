package slave

var N_BUTTONS = 3
var N_FLOORS = 4

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ElevatorDirection int

const (
	D_Down ElevatorDirection = -1
	D_Stop                   = 0
	D_Up                     = 1
)

type Elevator struct {
	floor     int
	direction ElevatorDirection
	requests  [N_FLOORS][N_BUTTONS]bool
	behaviour ElevatorBehaviour
}
