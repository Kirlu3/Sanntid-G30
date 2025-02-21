package master

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type HRAElevState struct {
	Floor       int                  `json:"floor"`
	Behavior    string               `json:"behaviour"`
	Direction   string               `json:"direction"`
	CabRequests [config.N_FLOORS]bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [config.N_FLOORS][2]bool `json:"hallRequests"` // first bool is for up and second is down
	States       map[string]HRAElevState `json:"states"`
}

var behaviorMap = map[slave.ElevatorBehaviour]string{
	slave.EB_Idle:     "idle",
	slave.EB_Moving:   "moving",
	slave.EB_DoorOpen: "doorOpen",
}

var directionMap = map[slave.ElevatorDirection]string{
	slave.D_Down: "down",
	slave.D_Stop: "stop",
	slave.D_Up:   "up",
}

func assignOrders(stateToAssign chan slave.WorldView, orderAssignments chan map[string][config.N_FLOORS][2]bool) {
	for {
		select {
		case state := <-stateToAssign:
			fmt.Printf("state: %v\n", state)

			assignments := assign(state)

			orderAssignments <- assignments
		}
	}
}

func assign(state slave.WorldView) map[string][config.N_FLOORS][2]bool {

	hraExecutable := ""

	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	input := transformInput(state) // transforms input from worldview to HRAInput

	// assign and returns output in json format
	outputJsonFormat, errAssign := exec.Command("./Project-resources/cost_fns/hall_request_assigner/"+hraExecutable, "-i", string(input)).CombinedOutput()

	if errAssign != nil {
		fmt.Println("Error occured when assigning: ", errAssign)
	}

	// transforms output from json format to the correct ouputformat
	output := transformOutput(outputJsonFormat)

	return output
}

func transformInput(state slave.WorldView) []byte { // transforms from WorldView to json format

	input := HRAInput{
		HallRequests: state.HallCalls,
		States:       map[string]HRAElevState{},
	}

	// adding all non-stuck and alive elevators to the state map
	for i := 0; i < len(state.Elevators); i++ {
		if !state.Elevators[i].Stuck && state.AliveElevators[i] {
			input.States[state.Elevators[i].Id] = HRAElevState{
				Floor:       state.Elevators[i].Floor,
				Behavior:    behaviorMap[state.Elevators[i].Behaviour],
				Direction:   directionMap[state.Elevators[i].Direction],
				CabRequests: state.CabCalls[i],
			}
		}
	}

	inputJsonFormat, errMarsial := json.Marshal(input)

	if errMarsial != nil {
		fmt.Println("Error using json.Marshal: ", errMarsial)
	}

	return inputJsonFormat
}

func transformOutput(outputJsonFormat []byte) map[string][config.N_FLOORS][2]bool {

	outputRightFormat := new(map[string][config.N_FLOORS][2]bool)

	errUnmarshal := json.Unmarshal(outputJsonFormat, &outputRightFormat)

	if errUnmarshal != nil {
		fmt.Println("Error using json.Unmarshal: ", errUnmarshal)
	}

	return *outputRightFormat
}