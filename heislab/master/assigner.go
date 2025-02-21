package master

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"` // first bool is for up and second is down
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

func assignOrders(stateToAssign <-chan slave.WorldView, toSlaveCh chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool, callsToAssign <-chan slave.Calls) {
	var state slave.WorldView
	for {
		select {
		case state = <-stateToAssign: // as far as assignOrders is concerned it doesnt matter if this comes directly from slaves or through stateManager
			fmt.Println("As:Received new states")
		default:
			select {
			case calls := <-callsToAssign:
				state.CabCalls = calls.CabCalls
				state.HallCalls = calls.HallCalls

				fmt.Printf("state: %v\n", state)
				assignments := assign(state)
				fmt.Printf("assigned:%v\n", assignments)
				toSlaveCh <- assignments
				fmt.Println("As:Succeded")
			default:
			}
		}
	}

}

func assign(state slave.WorldView) [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool { // [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool

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
	outputJsonFormat, errAssign := exec.Command("heislab/Project-resources/cost_fns/hall_request_assigner/"+hraExecutable, "-i", string(input)).CombinedOutput()

	if errAssign != nil {
		fmt.Println("Error occured when assigning: ", errAssign)
	}

	// transforms output from json format to the correct ouputformat
	output := transformOutput(outputJsonFormat, state)

	return output
}

func transformInput(state slave.WorldView) []byte { // transforms from WorldView to json format

	input := HRAInput{
		HallRequests: state.HallCalls[:],
		States:       map[string]HRAElevState{},
	}

	// adding all non-stuck and alive elevators to the state map
	for i := 0; i < len(state.Elevators); i++ {
		if !state.Elevators[i].Stuck && state.AliveElevators[i] {
			input.States[strconv.Itoa(state.Elevators[i].ID)] = HRAElevState{
				Floor:       state.Elevators[i].Floor,
				Behavior:    behaviorMap[state.Elevators[i].Behaviour],
				Direction:   directionMap[state.Elevators[i].Direction],
				CabRequests: state.CabCalls[i][:],
			}
		}
	}

	inputJsonFormat, errMarsial := json.Marshal(input)

	fmt.Println("Json input: ", input)

	if errMarsial != nil {
		fmt.Println("Error using json.Marshal: ", errMarsial)
	}

	return inputJsonFormat
}

func transformOutput(outputJsonFormat []byte, state slave.WorldView) [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool {
	output := [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool{}
	tempOutput := new(map[string][config.N_FLOORS][2]bool)

	errUnmarshal := json.Unmarshal(outputJsonFormat, &tempOutput)

	if errUnmarshal != nil {
		fmt.Println("Error using json.Unmarshal: ", errUnmarshal)
	}

	for elevatorKey, tempElevatorOrders := range *tempOutput {
		elevatorNr, err_convert := strconv.Atoi(elevatorKey)

		elevatorOrders := [config.N_FLOORS][config.N_BUTTONS]bool{}

		if err_convert != nil {
			fmt.Println("Error occured when converting to right assign format: ", err_convert)
		}

		for floor := range config.N_FLOORS {
			// appending cab calls from worldview of each floor to the output
			elevatorOrders[floor] = [3]bool{tempElevatorOrders[floor][0], tempElevatorOrders[floor][1], state.CabCalls[elevatorNr][floor]}
		}

		output[elevatorNr] = elevatorOrders
	}

	return output
}

/*

- assigner håndterer bare hall rqeuests og ikke cab requests

*/

/* struktur på tempOutput:
	"id_1" : [[Boolean, Boolean], ...],
    "id_2" : ...
*/

// struktur på output:
/*[
	[ elevator 0
		[up, down, cab], // floor 0
		[up, down, cab], // floor 1
		[up, down, cab], // floor 2
		[up, down, cab] // floor 3
	],
	[ elevator 1
		[[up, down, cab]],
		[[up, down, cab]],
		[[up, down, cab]],
		[[up, down, cab]]
	],
	[ elevator 2
		[[up, down, cab]],
		[[up, down, cab]],
		[[up, down, cab]],
		[[up, down, cab]]
	]
]
*/
