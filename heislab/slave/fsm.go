package Slave

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

func validElevator(elevator Elevator) bool {
	return elevator.behaviour > -1 && elevator.behaviour < 3 && //behaviour in bounds
		elevator.direction > -2 && elevator.direction < 2 && //direction in bounds
		elevator.floor > -1 && elevator.floor < N_FLOORS && //floor in bounds
		!elevator.requests[N_FLOORS-1][elevio.BT_HallUp] && !elevator.requests[0][elevio.BT_HallDown] && //no impossible requests
		!(elevator.behaviour == 2 && elevator.direction == 0) //no behaviour moving without direction
}

func fsm_onInit(elevator Elevator) Elevator {
	fmt.Println("onInit")
	elevator.direction = D_Down
	elevator.behaviour = EB_Moving
	fmt.Println("offInit")
	return elevator
}

func fsm_onRequestButtonPress(buttonEvent elevio.ButtonEvent, elevator Elevator) Elevator {
	fmt.Println("onRequestButtonPress")
	switch elevator.behaviour {
	case EB_DoorOpen:
		if !requests_shouldClearImmediately(elevator, buttonEvent.Floor, buttonEvent.Button) {
			elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
		}
	case EB_Moving:
		elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
	case EB_Idle:
		elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
		var pair DirectionBehaviourPair = requests_chooseDirection(elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour
		switch elevator.behaviour {
		case EB_DoorOpen:
			elevator = requests_clearAtCurrentFloor(elevator)
		}
	}
	return elevator
}

func fsm_onFloorArrival(newFloor int, elevator Elevator) Elevator {
	fmt.Println("onFloorArrival")
	elevator.floor = newFloor
	switch elevator.behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) {
			elevator = requests_clearAtCurrentFloor(elevator)
			elevator.behaviour = EB_DoorOpen
		}
	}
	return elevator
}

// not implemented yet
func fsm_onObstruction(obstruction bool) {
	if obstruction {
		fmt.Println("onObstruction -> True")
	} else {
		fmt.Println("onObstruction -> False")
	}
}

func fsm_onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
}

func fsm_onTimerEnd(elevator Elevator) Elevator {
	fmt.Println("onTimerEnd")

	switch elevator.behaviour {
	case EB_DoorOpen:
		var pair DirectionBehaviourPair = requests_chooseDirection(elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour
	}
	return elevator
}
