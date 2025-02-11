package Slave

import "github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"

func activateIO(n_elevator Elevator, elevator Elevator, t_start chan bool) {

	elevio.SetFloorIndicator(n_elevator.Floor) //Floor IO

	setAllLights(n_elevator) //Indicator lights IO

	if n_elevator.Behaviour != elevator.Behaviour {
		switch n_elevator.Behaviour {
		case EB_DoorOpen:
			t_start <- true
			elevio.SetDoorOpenLamp(true)
			elevio.SetMotorDirection(elevio.MD_Stop)
		case EB_Moving:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(n_elevator.Direction))
		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MD_Stop)
		}
	}
}

func setAllLights(elevator Elevator) {
	for Floor := 0; Floor < N_FLOORS; Floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), Floor, elevator.Requests[Floor][btn])
		}
	}
}
