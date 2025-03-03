package slave

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

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

// Checks if the new elevator object is within the bounds of the elevator system
func elevator_validElevator(elevator Elevator) bool {
	return elevator.Behaviour >= EB_Idle && elevator.Behaviour <= EB_Moving && //Behaviour in bounds
		elevator.Direction >= D_Down && elevator.Direction <= D_Up && //Direction in bounds
		elevator.Floor > -1 && elevator.Floor < config.N_FLOORS && //Floor in bounds
		!elevator.Requests[config.N_FLOORS-1][elevio.BT_HallUp] && !elevator.Requests[0][elevio.BT_HallDown] && //no impossible Requests
		!(elevator.Behaviour == EB_Moving && elevator.Direction == D_Stop) //no Behaviour moving without Direction
}

// Not a pure function as it also activates IO, I think this is fine though
func elevator_updateElevator(n_elevator Elevator, elevator Elevator, tx chan<- EventMessage, t_start chan int) Elevator {
	if elevator_validElevator(n_elevator) {
		if n_elevator.Stuck != elevator.Stuck { //if stuck status has changed
			tx <- EventMessage{0, n_elevator, Stuck, elevio.ButtonEvent{}} //send message to master
		}
		io_activateIO(n_elevator, elevator, t_start)
		return n_elevator
	} else {
		panic("Invalid elevator")
	}
}
