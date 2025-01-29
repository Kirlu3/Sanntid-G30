package slave

type DirectionBehaviorPair struct {
	direction ElevatorDirection
	behavior  ElevatorBehaviour
}

func requests_above(elevator Elevator) bool {
	for f := elevator.floor; f < N_FLOORS; f++ {
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

func requests_chooseDirection(elevator Elevator) DirectionBehaviorPair {
	switch elevator.direction {
	case D_Up:
		if requests_above(elevator) {
			return DirectionBehaviorPair{D_Up, EB_Moving}
		} else if request_here(elevator) {
			return DirectionBehaviorPair{D_Down, EB_DoorOpen}
		} else if requests_below(elevator) {
			return DirectionBehaviorPair{D_Down, EB_Moving}
		} else {
			return DirectionBehaviorPair{D_Stop, EB_Idle}
		}
	case D_Down:
		if requests_below(elevator) {
			return DirectionBehaviorPair{D_Down, EB_Moving}
		} else if request_here(elevator) {
			return DirectionBehaviorPair{D_Up, EB_DoorOpen}
		} else if requests_above(elevator) {
			return DirectionBehaviorPair{D_Up, EB_Moving}
		} else {
			return DirectionBehaviorPair{D_Stop, EB_Idle}
		}
	case D_Stop:
		if request_here(elevator) {
			return DirectionBehaviorPair{D_Stop, EB_DoorOpen}
		} else if requests_above(elevator) {
			return DirectionBehaviorPair{D_Up, EB_Moving}
		} else if requests_below(elevator) {
			return DirectionBehaviorPair{D_Down, EB_Moving}
		} else {
			return DirectionBehaviorPair{D_Stop, EB_Idle}
		}
	}
	return DirectionBehaviorPair{D_Stop, EB_Idle}
}
