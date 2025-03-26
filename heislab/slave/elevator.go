package slave

import (
	"time"

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
	ID          int
	Floor       int
	Direction   ElevatorDirection
	Calls       [config.N_FLOORS][config.N_BUTTONS]bool
	Behaviour   ElevatorBehaviour
	Stuck       bool
	Obstruction bool
}

/*
# Checks if an elevator has a state within the correct bounds

Input: The elevator to be checked

Returns: True if the elevator is valid, false otherwise
*/
func validElevator(elevator Elevator) bool {
	return elevator.Behaviour >= EB_Idle && elevator.Behaviour <= EB_Moving && //Behaviour in bounds
		elevator.Direction >= D_Down && elevator.Direction <= D_Up && //Direction in bounds
		elevator.Floor > -1 && elevator.Floor < config.N_FLOORS && //Floor in bounds
		!elevator.Calls[config.N_FLOORS-1][elevio.BT_HallUp] && !elevator.Calls[0][elevio.BT_HallDown] && //no impossible Calls
		!(elevator.Behaviour == EB_Moving && elevator.Direction == D_Stop) //no Behaviour moving without Direction
}

/*
# Updates the state of elevator with the state of newElevator if its is valid, sends the new state to the braodcaster, and starts the correct timer.

Input: The new state of the elevator, the current state of the elevator, the channel to send the new state to the broadcaster, and the channel to send the timer duration to the timer.

Returns: The updated elevator struct
*/
func updateElevatorState(newElevator Elevator, elevator Elevator, slaveStateToMasterChan chan<- Elevator, timer *time.Timer) Elevator {
	if validElevator(newElevator) {
		slaveStateToMasterChan <- newElevator

		switch newElevator.Behaviour {
		case EB_DoorOpen:
			if newElevator.Calls != elevator.Calls || newElevator.Obstruction {
				timer.Reset(time.Second * time.Duration(config.DoorOpenDuration))
			}
		case EB_Moving:
			timer.Reset(time.Second * time.Duration(config.TimeBetweenFloors))
		}

		return newElevator
	}
	return elevator
}

/*
Returns true if there are calls above the elevator's current floor

Else returns false
*/
func callsAboveElevator(elevator Elevator) bool {
	for f := elevator.Floor + 1; f < config.N_FLOORS; f++ {
		for btn := range config.N_BUTTONS {
			if elevator.Calls[f][btn] {
				return true
			}
		}
	}
	return false
}

/*
Returns true if there are calls below the elevator's current floor

Else returns false
*/
func callsBelowElevator(elevator Elevator) bool {
	for f := range elevator.Floor {
		for btn := range config.N_BUTTONS {
			if elevator.Calls[f][btn] {
				return true
			}
		}
	}
	return false
}

/*
Returns true if there are calls at the elevator's current floor

Else returns false
*/
func callsAtCurrentFloor(elevator Elevator) bool {
	for btn := range config.N_BUTTONS {
		if elevator.Calls[elevator.Floor][btn] {
			return true
		}
	}
	return false
}

/*
# Chooses the direction and behaviour of the elevator based on the calls and current state of the elevator

Input: The elevator

Returns: The chosen direction and behaviour for the elevator
*/
func chooseElevatorDirection(elevator Elevator) (ElevatorDirection, ElevatorBehaviour) {
	switch elevator.Direction {
	case D_Up:
		if elevator.Calls[elevator.Floor][elevio.BT_HallUp] || elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return D_Up, EB_DoorOpen
		} else if callsAboveElevator(elevator) {
			return D_Up, EB_Moving
		} else if callsAtCurrentFloor(elevator) {
			return D_Down, EB_DoorOpen
		} else if callsBelowElevator(elevator) {
			return D_Down, EB_Moving
		} else {
			return D_Up, EB_Idle
		}
	case D_Down:
		if elevator.Calls[elevator.Floor][elevio.BT_HallDown] || elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return D_Down, EB_DoorOpen
		} else if callsBelowElevator(elevator) {
			return D_Down, EB_Moving
		} else if callsAtCurrentFloor(elevator) {
			return D_Up, EB_DoorOpen
		} else if callsAboveElevator(elevator) {
			return D_Up, EB_Moving
		} else {
			return D_Down, EB_Idle
		}
	case D_Stop:
		if callsAtCurrentFloor(elevator) {
			return D_Stop, EB_DoorOpen
		} else if callsAboveElevator(elevator) {
			return D_Up, EB_Moving
		} else if callsBelowElevator(elevator) {
			return D_Down, EB_Moving
		} else {
			return D_Stop, EB_Idle
		}
	default:
		panic("Invalid elevator direction")
	}
}

/*
Returns true if the elevator should stop at the current floor

Else returns false
*/
func shouldElevatorStop(elevator Elevator) bool {
	switch elevator.Direction {
	case D_Down:
		if elevator.Calls[elevator.Floor][elevio.BT_HallDown] {
			return true
		}
		if elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return true
		}
		if !callsBelowElevator(elevator) {
			return true
		}
		return false
	case D_Up:
		if elevator.Calls[elevator.Floor][elevio.BT_HallUp] {
			return true
		}
		if elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return true
		}
		if !callsAboveElevator(elevator) {
			return true
		}
		return false
	default:
		return true
	}
}

/*
# Clears calls depending on the direction of the elevator. Will only clear hall calls in one direction at a time.

Input: the elevator

Returns: the elevator with cleared calls at the current floor, in addition to a new direction in the case of no initial direction.
*/
func clearCallsAtCurrentFloor(elevator Elevator) Elevator {
	elevator.Calls[elevator.Floor][elevio.BT_Cab] = false
	switch elevator.Direction {
	case D_Up:
		elevator.Calls[elevator.Floor][elevio.BT_HallUp] = false
	case D_Down:
		elevator.Calls[elevator.Floor][elevio.BT_HallDown] = false
	default:
		if elevator.Calls[elevator.Floor][elevio.BT_HallUp] {
			elevator.Calls[elevator.Floor][elevio.BT_HallUp] = false
			elevator.Direction = D_Up
		} else if elevator.Calls[elevator.Floor][elevio.BT_HallDown] {
			elevator.Calls[elevator.Floor][elevio.BT_HallDown] = false
			elevator.Direction = D_Down
		}
	}
	return elevator
}
