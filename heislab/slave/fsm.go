package Slave

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

func validElevator(elevator Elevator) bool {
	return elevator.Behaviour > -1 && elevator.Behaviour < 3 && //Behaviour in bounds
		elevator.Direction > -2 && elevator.Direction < 2 && //Direction in bounds
		elevator.Floor > -1 && elevator.Floor < N_FLOORS && //Floor in bounds
		!elevator.Requests[N_FLOORS-1][elevio.BT_HallUp] && !elevator.Requests[0][elevio.BT_HallDown] && //no impossible Requests
		!(elevator.Behaviour == 2 && elevator.Direction == 0) //no Behaviour moving without Direction
}

func fsm_onInit(elevator Elevator) Elevator {
	fmt.Println("onInit")
	elevator.Direction = D_Down
	elevator.Behaviour = EB_Moving
	fmt.Println("offInit")
	return elevator
}

func fsm_onRequestButtonPress(buttonEvent elevio.ButtonEvent, elevator Elevator) Elevator {
	fmt.Println("onRequestButtonPress")
	switch elevator.Behaviour {
	case EB_DoorOpen:
		if !Requests_shouldClearImmediately(elevator, buttonEvent.Floor, buttonEvent.Button) {
			elevator.Requests[buttonEvent.Floor][buttonEvent.Button] = true
		}
	case EB_Moving:
		elevator.Requests[buttonEvent.Floor][buttonEvent.Button] = true
	case EB_Idle:
		elevator.Requests[buttonEvent.Floor][buttonEvent.Button] = true
		var pair DirectionBehaviourPair = Requests_chooseDirection(elevator)
		elevator.Direction = pair.Direction
		elevator.Behaviour = pair.Behaviour
		switch elevator.Behaviour {
		case EB_DoorOpen:
			elevator = Requests_clearAtCurrentFloor(elevator)
		}
	}
	return elevator
}

func fsm_onFloorArrival(newFloor int, elevator Elevator) Elevator {
	elevator.Stuck = false //if the elevator arrives at a floor, it is not stuck
	fmt.Println("onFloorArrival")
	elevator.Floor = newFloor
	switch elevator.Behaviour {
	case EB_Moving:
		if Requests_shouldStop(elevator) {
			elevator = Requests_clearAtCurrentFloor(elevator)
			elevator.Behaviour = EB_DoorOpen
		}
	}
	return elevator
}

// not implemented yet
func fsm_onObstruction(obstruction bool, elevator Elevator) Elevator {
	elevator.Stuck = obstruction
	return elevator
}

func fsm_onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
}

func fsm_onTimerEnd(elevator Elevator) Elevator {
	fmt.Println("onTimerEnd")

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if !elevator.Stuck {
			var pair DirectionBehaviourPair = Requests_chooseDirection(elevator)
			elevator.Direction = pair.Direction
			elevator.Behaviour = pair.Behaviour
		}
	case EB_Moving:
		elevator.Stuck = true
	}
	return elevator
}
