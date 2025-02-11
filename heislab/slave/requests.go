package Slave

import "github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"

type DirectionBehaviourPair struct {
	Direction ElevatorDirection
	Behaviour ElevatorBehaviour
}

func Requests_above(elevator Elevator) bool {
	for f := elevator.Floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func Requests_below(elevator Elevator) bool {
	for f := 0; f < elevator.Floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func request_here(elevator Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if elevator.Requests[elevator.Floor][btn] {
			return true
		}
	}
	return false
}

func Requests_chooseDirection(elevator Elevator) DirectionBehaviourPair {
	switch elevator.Direction {
	case D_Up:
		if Requests_above(elevator) {
			return DirectionBehaviourPair{D_Up, EB_Moving}
		} else if request_here(elevator) {
			return DirectionBehaviourPair{D_Down, EB_DoorOpen}
		} else if Requests_below(elevator) {
			return DirectionBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirectionBehaviourPair{D_Stop, EB_Idle}
		}
	case D_Down:
		if Requests_below(elevator) {
			return DirectionBehaviourPair{D_Down, EB_Moving}
		} else if request_here(elevator) {
			return DirectionBehaviourPair{D_Up, EB_DoorOpen}
		} else if Requests_above(elevator) {
			return DirectionBehaviourPair{D_Up, EB_Moving}
		} else {
			return DirectionBehaviourPair{D_Stop, EB_Idle}
		}
	case D_Stop:
		if request_here(elevator) {
			return DirectionBehaviourPair{D_Stop, EB_DoorOpen}
		} else if Requests_above(elevator) {
			return DirectionBehaviourPair{D_Up, EB_Moving}
		} else if Requests_below(elevator) {
			return DirectionBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirectionBehaviourPair{D_Stop, EB_Idle}
		}
	}
	return DirectionBehaviourPair{D_Stop, EB_Idle}
}

func Requests_shouldStop(elevator Elevator) bool {
	switch elevator.Direction {
	case D_Down:
		if elevator.Requests[elevator.Floor][elevio.BT_HallDown] {
			return true
		}
		if elevator.Requests[elevator.Floor][elevio.BT_Cab] {
			return true
		}
		if !Requests_below(elevator) {
			return true
		}
		return false
	case D_Up:
		if elevator.Requests[elevator.Floor][elevio.BT_HallUp] {
			return true
		}
		if elevator.Requests[elevator.Floor][elevio.BT_Cab] {
			return true
		}
		if !Requests_above(elevator) {
			return true
		}
		return false
	default:
		return true
	}
}

func Requests_shouldClearImmediately(elevator Elevator, btn_Floor int, btn_type elevio.ButtonType) bool {
	//only case CV_InDirn from C:
	return elevator.Floor == btn_Floor && ((elevator.Direction == D_Up && btn_type == elevio.BT_HallUp) ||
		(elevator.Direction == D_Down && btn_type == elevio.BT_HallDown) ||
		elevator.Direction == D_Stop ||
		btn_type == elevio.BT_Cab)
}

func Requests_clearAtCurrentFloor(elevator Elevator) Elevator {
	//only case CV_InDirn from C:
	elevator.Requests[elevator.Floor][elevio.BT_Cab] = false
	switch elevator.Direction {
	case D_Up:
		if !Requests_above(elevator) && !elevator.Requests[elevator.Floor][elevio.BT_HallUp] {
			elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
		}
		elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
	case D_Down:
		if !Requests_below(elevator) && !elevator.Requests[elevator.Floor][elevio.BT_HallDown] {
			elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
		}
		elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
	default:
		elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
		elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
	}
	return elevator
}
