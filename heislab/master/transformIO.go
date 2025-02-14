package Master

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

var BehaviorMap = map[Slave.ElevatorBehaviour]string{
	Slave.EB_Idle: "idle", 
	Slave.EB_Moving: "moving", 
	Slave.EB_DoorOpen: "doorOpen"}


func transformInput(elevators []Slave.Elevator, hallRequests [][2]bool)HRAInput{
	// transforms from Elevator struct to HRAInput struct

	// HRAInput struct
	input := HRAInput{
		HallRequests: hallRequests,
		States: map[string]HRAElevState{},
	}	

	// adding all elevators to the state map
	for i := 0; i < len(elevators); i++{
		input.States[string(i)] = HRAElevState{
			Floor: elevators[i].floor,
			Behavior: BehaviorMap[elevators[i].behaviour],
			Direction: string(elevators[i].direction),
			CabRequests: elevators[i].requests,
			Obstruction: false, // have not been implemented yet
		}
	}

	return input
}	

func transformOutput(){
	// convert output to right format 
}



// notes:
// - må vi legge til egen liste med cab requests i Elevator structen?