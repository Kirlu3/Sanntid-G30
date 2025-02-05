package Slave

import "github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"

func activateIO(n_elevator Elevator, elevator Elevator, t_start chan bool) {

	elevio.SetFloorIndicator(n_elevator.floor) //Floor IO

	setAllLights(n_elevator) //Indicator lights IO

	if n_elevator.behaviour != elevator.behaviour {
		switch n_elevator.behaviour {
		case EB_DoorOpen:
			t_start <- true
			elevio.SetDoorOpenLamp(true)
			elevio.SetMotorDirection(elevio.MD_Stop)
		case EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(n_elevator.direction))
		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MD_Stop)
		}
	}
}

func setAllLights(elevator Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.requests[floor][btn])
		}
	}
}
