package slave

import (
	"fmt"

	"../driver-go/elevio"
)

var ob = make(chan bool)

var elevator Elevator

func setAllLights(es Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, es.requests[floor][btn])
		}
	}
}

func fsm_onInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = D_Down
	elevator.behaviour = EB_Moving
}

func fsm_onRequestButtonPress(buttonEvent elevio.ButtonEvent, t_start chan bool) {
	fmt.Println("onRequestButtonPress")
	switch elevator.behaviour {
	case EB_DoorOpen:
		if requests_shouldClearImmediately(elevator, buttonEvent.Floor, buttonEvent.Button) {
			//restart door timer
			t_start <- true
		} else {
			elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
		}
		//Do nothing
	case EB_Moving:
		elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
	case EB_Idle:
		elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
		var pair DirectionBehaviorPair = requests_chooseDirection(elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behavior
		switch elevator.behaviour {
		case EB_DoorOpen:
			//Open door
			t_start <- true
			elevio.SetDoorOpenLamp(true)
			elevator = requests_clearAtCurrentFloor(elevator)
		case EB_Moving:
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		case EB_Idle:
		}
		elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		return
	}
}

func fsm_onFloorArrival(newFloor int, t_start chan bool) {
	fmt.Println("onFloorArrival")
	elevator.floor = newFloor
	elevio.SetFloorIndicator(elevator.floor)
	switch elevator.behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) {
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
			elevio.SetDoorOpenLamp(true)
			elevator = requests_clearAtCurrentFloor(elevator)
			t_start <- true
			setAllLights(elevator)
			elevator.behaviour = EB_DoorOpen
		}
	}
}

// not implemented yet? This is an attempt that might easily make bugs
func fsm_onObstruction(obstruction bool) {
	<-ob
	ob <- obstruction
	return
}

func fsm_onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
	return
}

func fsm_onTimerEnd(t_start chan bool) {

	for <-ob {
		ob <- true
	}
	ob <- false

	switch elevator.behaviour {
	case EB_DoorOpen:
		var pair DirectionBehaviorPair = requests_chooseDirection(elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behavior

		switch elevator.behaviour {
		case EB_DoorOpen:
			t_start <- true
			elevator = requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Moving:
			//no way this is correct?
		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		}

	}

}
