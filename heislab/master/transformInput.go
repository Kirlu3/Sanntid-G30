package master

import (
	"encoding/json"
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type HRAElevState struct {
    Floor       int         `json:"floor"` 
    Behavior    string      `json:"behaviour"`
    Direction   string      `json:"direction"`
    CabRequests []bool      `json:"cabRequests"`
	Stuck bool				`json:"stuck"`
}

type HRAInput struct {
    HallRequests    [][2]bool                   `json:"hallRequests"` // first bool is for up and second is down
    States          map[string]HRAElevState     `json:"states"`
}

var BehaviorMap = map[slave.ElevatorBehaviour]string{
	slave.EB_Idle: "idle", 
	slave.EB_Moving: "moving", 
	slave.EB_DoorOpen: "doorOpen",
}

var DirectionMap = map[slave.ElevatorDirection]string{
	slave.D_Down : "down",
	slave.D_Stop : "stop",
	slave.S_Up : "up",
}


func transformInput(state slave.WordlView)[]byte{
	// transforms from Elevator struct to HRAInput struct

	// HRAInput struct
	input := HRAInput{
		HallRequests: state.HallCalls,
		States: map[string]HRAElevState{},
	}	

	// adding all elevators to the state map
	for i := 0; i < len(state.Elevators); i++{
		input.States[state.Elevators[i].Id] = HRAElevState{
			Floor: state.Elevators[i].Floor,
			Behavior: BehaviorMap[state.Elevators[i].Behavior],
			Direction: DirectionMap[state.Elevators[i].Direction],
			CabRequests: state.CabCalls[i],
			Stuck: state.Elevators[i].Stuck, // have not been implemented yet
		}
	}

	
	// makes input into json format 
	inputJsonFormat, errMarsial := json.Marshal(input)

	if errMarsial != nil{
		fmt.Println("Error using json.Marshal: ", errMarsial)
	}

	return inputJsonFormat
}	


func transformOutput(outputJsonFormat []byte)map[string][slave.N_FLOORS][2]bool{

	
	outputRightFormat := new(map[string][slave.N_FLOORS][2]bool)

	errUnmarshal := json.Unmarshal(outputJsonFormat, &outputRightFormat)

	if errUnmarshal != nil {
		fmt.Println("Error using json.Unmarshal: ", errUnmarshal)
	}

	return outputRightFormat
}