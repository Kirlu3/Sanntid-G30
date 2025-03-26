package slave

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

/*
# Activates the elevator IO at the start of each FSM loop iteration. Interfaces with the elevator hardware to update the lights and motor direction based on elevator state.

Input: The most recent state of the elevator.
*/
func activateElevatorIO(elevator Elevator) {

	elevio.SetFloorIndicator(elevator.Floor)

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
# Updates the hall and cab call lights by interfacing with the elevator hardware.

Input: Array of how the lights should be updated.
*/
func updateLights(lights [config.N_FLOORS][config.N_BUTTONS]bool) {
	for i := range config.N_FLOORS {
		for j := range config.N_BUTTONS {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, lights[i][j])
		}
	}
}
