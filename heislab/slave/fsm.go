package Slave

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

var elevator Elevator

func setAllLights(es Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, es.requests[floor][btn])
		}
	}
}

func fsm_onInit() {
	fmt.Println("onInit")
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = D_Down
	elevator.behaviour = EB_Moving
	fmt.Println("offInit")
}

func fsm_onRequestButtonPress(buttonEvent elevio.ButtonEvent, t_start chan bool) {
	fmt.Println("onRequestButtonPress")
	switch elevator.behaviour {
	case EB_DoorOpen:
		if requests_shouldClearImmediately(elevator, buttonEvent.Floor, buttonEvent.Button) {
			fmt.Println("Starting timer")
			t_start <- true
			fmt.Println("Started timer")
		} else {
			elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
		}
		//Do nothing
	case EB_Moving:
		elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
	case EB_Idle:
		elevator.requests[buttonEvent.Floor][buttonEvent.Button] = true
		var pair DirectionBehaviourPair = requests_chooseDirection(elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour
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
	}
	setAllLights(elevator)
}

func fsm_onFloorArrival(newFloor int, t_start chan bool) {
	fmt.Println("onFloorArrival")
	elevator.floor = newFloor
	elevio.SetFloorIndicator(elevator.floor)
	switch elevator.behaviour {
	case EB_Moving:
		if requests_shouldStop(elevator) {
			elevio.SetMotorDirection(elevio.MD_Stop)
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
	if obstruction {
		fmt.Println("onObstruction -> True")
	} else {
		fmt.Println("onObstruction -> False")
	}
}

func fsm_onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
}

func fsm_onTimerEnd(t_start chan bool) {
	fmt.Println("onTimerEnd")

	switch elevator.behaviour {
	case EB_DoorOpen:
		var pair DirectionBehaviourPair = requests_chooseDirection(elevator)
		elevator.direction = pair.direction
		elevator.behaviour = pair.behaviour

		switch pair.behaviour {
		case EB_DoorOpen:
			t_start <- true
			elevator = requests_clearAtCurrentFloor(elevator)
			setAllLights(elevator)
		case EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		}
	default:
		return
	}
}
