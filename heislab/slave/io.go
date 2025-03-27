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
	case BehaviourDoorOpen:
		elevio.SetDoorOpenLamp(true)
		elevio.SetMotorDirection(elevio.MDirectionStop)
	case BehaviourMoving:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MotorDirection(elevator.Direction))
	case BehaviourIdle:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MDirectionStop)
	}
}

/*
# Updates the hall and cab call lights by interfacing with the elevator hardware.

Input: Array of how the lights should be updated.
*/
func updateLights(lights [config.NumFloors][config.NumBtns]bool) {
	for i := range config.NumFloors {
		for j := range config.NumBtns {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, lights[i][j])
		}
	}
}
