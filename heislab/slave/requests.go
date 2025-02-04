package Slave

import "github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"

type DirectionBehaviourPair struct {
	direction ElevatorDirection
	behaviour ElevatorBehaviour
}

func requests_above(elevator Elevator) bool {
	for f := elevator.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(elevator Elevator) bool {
	for f := 0; f < elevator.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func request_here(elevator Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if elevator.requests[elevator.floor][btn] {
			return true
		}
	}
	return false
}

func requests_chooseDirection(elevator Elevator) DirectionBehaviourPair {
	switch elevator.direction {
	case D_Up:
		if requests_above(elevator) {
			return DirectionBehaviourPair{D_Up, EB_Moving}
		} else if request_here(elevator) {
			return DirectionBehaviourPair{D_Down, EB_DoorOpen}
		} else if requests_below(elevator) {
			return DirectionBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirectionBehaviourPair{D_Stop, EB_Idle}
		}
	case D_Down:
		if requests_below(elevator) {
			return DirectionBehaviourPair{D_Down, EB_Moving}
		} else if request_here(elevator) {
			return DirectionBehaviourPair{D_Up, EB_DoorOpen}
		} else if requests_above(elevator) {
			return DirectionBehaviourPair{D_Up, EB_Moving}
		} else {
			return DirectionBehaviourPair{D_Stop, EB_Idle}
		}
	case D_Stop:
		if request_here(elevator) {
			return DirectionBehaviourPair{D_Stop, EB_DoorOpen}
		} else if requests_above(elevator) {
			return DirectionBehaviourPair{D_Up, EB_Moving}
		} else if requests_below(elevator) {
			return DirectionBehaviourPair{D_Down, EB_Moving}
		} else {
			return DirectionBehaviourPair{D_Stop, EB_Idle}
		}
	}
	return DirectionBehaviourPair{D_Stop, EB_Idle}
}

func requests_shouldStop(elevator Elevator) bool {
	switch elevator.direction {
	case D_Down:
		if elevator.requests[elevator.floor][elevio.BT_HallDown] {
			return true
		}
		if elevator.requests[elevator.floor][elevio.BT_Cab] {
			return true
		}
		if !requests_below(elevator) {
			return true
		}
		return false
	case D_Up:
		if elevator.requests[elevator.floor][elevio.BT_HallUp] {
			return true
		}
		if elevator.requests[elevator.floor][elevio.BT_Cab] {
			return true
		}
		if !requests_above(elevator) {
			return true
		}
		return false
	default:
		return true
	}
}

func requests_shouldClearImmediately(elevator Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	//only case CV_InDirn from C:
	return elevator.floor == btn_floor && ((elevator.direction == D_Up && btn_type == elevio.BT_HallUp) ||
		(elevator.direction == D_Down && btn_type == elevio.BT_HallDown) ||
		elevator.direction == D_Stop ||
		btn_type == elevio.BT_Cab)
}

func requests_clearAtCurrentFloor(elevator Elevator) Elevator {
	//only case CV_InDirn from C:
	elevator.requests[elevator.floor][elevio.BT_Cab] = false //I doubt we can actually clear in this way
	switch elevator.direction {
	case D_Up:
		if !requests_above(elevator) && !elevator.requests[elevator.floor][elevio.BT_HallUp] {
			elevator.requests[elevator.floor][elevio.BT_HallDown] = false
		}
		elevator.requests[elevator.floor][elevio.BT_HallUp] = false
	case D_Down:
		if !requests_below(elevator) && !elevator.requests[elevator.floor][elevio.BT_HallDown] {
			elevator.requests[elevator.floor][elevio.BT_HallUp] = false
		}
		elevator.requests[elevator.floor][elevio.BT_HallDown] = false
	case D_Stop:
	default:
		elevator.requests[elevator.floor][elevio.BT_HallUp] = false
		elevator.requests[elevator.floor][elevio.BT_HallDown] = false
	}
	return elevator
}
