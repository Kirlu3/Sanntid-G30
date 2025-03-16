package slave

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

/*
Returns true if there are calls above the elevator's current floor

Else returns false
*/
func calls_above(elevator Elevator) bool {
	for f := elevator.Floor + 1; f < config.N_FLOORS; f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
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
func calls_below(elevator Elevator) bool {
	for f := 0; f < elevator.Floor; f++ {
		for btn := 0; btn < config.N_BUTTONS; btn++ {
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
func calls_here(elevator Elevator) bool {
	for btn := 0; btn < config.N_BUTTONS; btn++ {
		if elevator.Calls[elevator.Floor][btn] {
			return true
		}
	}
	return false
}

/*
	Chooses the direction and behaviour of the elevator based on the calls and current state

Input: Original elevator object

Returns: Direction and behaviour the elevator should have
*/
func calls_chooseDirection(elevator Elevator) (ElevatorDirection, ElevatorBehaviour) {
	switch elevator.Direction {
	case D_Up:
		if calls_above(elevator) {
			return D_Up, EB_Moving
		} else if calls_here(elevator) {
			return D_Down, EB_DoorOpen
		} else if calls_below(elevator) {
			return D_Down, EB_Moving
		} else {
			return D_Stop, EB_Idle
		}
	case D_Down:
		if calls_below(elevator) {
			return D_Down, EB_Moving
		} else if calls_here(elevator) {
			return D_Up, EB_DoorOpen
		} else if calls_above(elevator) {
			return D_Up, EB_Moving
		} else {
			return D_Stop, EB_Idle
		}
	case D_Stop:
		if calls_here(elevator) {
			return D_Stop, EB_DoorOpen
		} else if calls_above(elevator) {
			return D_Up, EB_Moving
		} else if calls_below(elevator) {
			return D_Down, EB_Moving
		} else {
			return D_Stop, EB_Idle
		}
	}
	return D_Stop, EB_Idle
}

/*
Returns true if the elevator should stop at the current floor

Else returns false
*/
func calls_shouldStop(elevator Elevator) bool {
	switch elevator.Direction {
	case D_Down:
		if elevator.Calls[elevator.Floor][elevio.BT_HallDown] {
			return true
		}
		if elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return true
		}
		if !calls_below(elevator) {
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
		if !calls_above(elevator) {
			return true
		}
		return false
	default:
		return true
	}
}

/*
Clears calls depending on the direction of the elevator

Input: Original elevator object

Returns: New elevator object with cleared calls
*/
func calls_clearAtCurrentFloor(elevator Elevator) Elevator {
	elevator.Calls[elevator.Floor][elevio.BT_Cab] = false
	switch elevator.Direction {
	case D_Up:
		elevator.Calls[elevator.Floor][elevio.BT_HallUp] = false
	case D_Down:
		elevator.Calls[elevator.Floor][elevio.BT_HallDown] = false
	default:
		elevator.Calls[elevator.Floor][elevio.BT_HallUp] = false
		elevator.Calls[elevator.Floor][elevio.BT_HallDown] = false
	}
	fmt.Println("Cleared at current floor:", elevator.Calls)
	return elevator
}
