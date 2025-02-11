package Slave

import "github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"

const doorOpenDuration = 3
const timeBetweenFloors = 5

func activateIO(n_elevator Elevator, elevator Elevator, t_start chan int) {

	elevio.SetFloorIndicator(n_elevator.Floor) //Floor IO

	if n_elevator.Behaviour != elevator.Behaviour {
		switch n_elevator.Behaviour {
		case EB_DoorOpen:
			t_start <- doorOpenDuration
			elevio.SetDoorOpenLamp(true)
			elevio.SetMotorDirection(elevio.MD_Stop)
		case EB_Moving:
			t_start <- timeBetweenFloors
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(n_elevator.Direction))
		case EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MD_Stop)
		}
	}
}
