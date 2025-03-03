package slave

import (
	"fmt"
)

func fsm_onInit(elevator Elevator) Elevator {
	fmt.Println("onInit")
	elevator.Direction = D_Down
	elevator.Behaviour = EB_Moving
	fmt.Println("offInit")
	return elevator
}

/*
	Activates when the elevator receives new requests

Input: the old elevator object with updated requests

Returns: the new elevator object with updated direction and behaviour
*/
func fsm_onRequests(elevator Elevator) Elevator {
	fmt.Println("onRequest")
	switch elevator.Behaviour {
	case EB_DoorOpen:
		elevator = requests_clearAtCurrentFloor(elevator)
	case EB_Moving:
	case EB_Idle:
		direction, behaviour := requests_chooseDirection(elevator)
		elevator.Direction = direction
		elevator.Behaviour = behaviour
		if elevator.Behaviour == EB_DoorOpen {
			elevator = requests_clearAtCurrentFloor(elevator)
		}
	}
	return elevator
}

/*
	Activates when the elevator floor sensor is triggered

Input: the new floor and the old elevator object

Returns: the new elevator object
*/
func fsm_onFloorArrival(newFloor int, elevator Elevator) Elevator {
	elevator.Stuck = false //if the elevator arrives at a floor, it is not stuck
	fmt.Println("onFloorArrival")
	elevator.Floor = newFloor
	switch elevator.Behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) { //This causes the door to open on init, probably fine?
			elevator = requests_clearAtCurrentFloor(elevator)
			elevator.Behaviour = EB_DoorOpen
		}
	}
	return elevator
}

/*
	Activates when the obstruction sensor is triggered

Input: the old elevator object with updated obstruction status and behaviour

Returns: the new elevator object
*/
func fsm_onObstruction(obstruction bool, elevator Elevator) Elevator {
	fmt.Println("onObstruction")
	elevator.Stuck = obstruction
	if obstruction {
		elevator.Behaviour = EB_DoorOpen
		elevator.Direction = D_Stop
	} else {
		direction, behaviour := requests_chooseDirection(elevator)
		elevator.Direction = direction
		elevator.Behaviour = behaviour
	}
	return elevator
}

/*
	Activates when the stop button sensor is triggered

Does nothing but print a message
*/
func fsm_onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
}

/*
	Activates when the timer ends
	Either the door should close or the elevator is stuck

Input: the old elevator object

Returns: the new elevator object
*/
func fsm_onTimerEnd(elevator Elevator) Elevator {

	switch elevator.Behaviour {
	case EB_DoorOpen:
		fmt.Println("FSM:onTimerEnd DO")

		if !elevator.Stuck {
			direction, behaviour := requests_chooseDirection(elevator)
			elevator.Direction = direction
			elevator.Behaviour = behaviour
		}
	case EB_Moving:
		fmt.Println("FSM:onTimerEnd M")
		elevator.Stuck = true
	}
	return elevator
}
