package slave

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

/*
	Activates when the elevator state is updated
	Interfaces with the elevator hardware to update the lights and motor direction
	If the door opens or the elevator starts moving, the corresponding timer is started

Input: The elevator with updated state and timerDurationChan that starts a timer with the specified duration in seconds.
*/
func activateElevatorIO(elevator Elevator) {

	elevio.SetFloorIndicator(elevator.Floor) //Floor IO

	switch elevator.Behaviour {
	case EB_DoorOpen:
		elevio.SetDoorOpenLamp(true)
		elevio.SetMotorDirection(elevio.MD_Stop)
	case EB_Moving:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MotorDirection(elevator.Direction))
	case EB_Idle:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}

/*
	Updates the lights on the elevator panel.
	Interfaces with the elevator hardware to update the lights.

Input: Array of how the lights should be updated.
*/
func updateLights(lights [config.N_FLOORS][config.N_BUTTONS]bool) {
	for i := range config.N_FLOORS {
		for j := range config.N_BUTTONS {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, lights[i][j])
		}
	}

}
