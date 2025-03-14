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

func assignOrders(
	stateUpdateCh <-chan slave.Elevator,
	callsToAssignCh <-chan AssignCalls,
	assignmentsToSlaveCh chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	assignmentsToSlaveReceiver chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
) {
	
	elevators := [config.N_ELEVATORS]slave.Elevator{} // consider waiting for state init
	calls := AssignCalls{}

	for i := range config.N_ELEVATORS {
		elevators[i].ID = i // suggested fix to assigner init bug
	}
	for {
		select {
		case stateUpdate := <-stateUpdateCh:
			prevElevator := elevators[stateUpdate.ID]
			elevators[stateUpdate.ID] = stateUpdate

			if prevElevator.Stuck != stateUpdate.Stuck { // reassign if elev has become stuck/unstuck
				calls.AliveElevators[stateUpdate.ID] = !stateUpdate.Stuck // acts as if the elevator is dead if it is stuck
				assignments := assign(elevators, calls)
				assignmentsToSlaveCh <- assignments
				assignmentsToSlaveReceiver <- assignments
			}
			fmt.Println("As:Received new states")
		default:
			select {
			case calls = <-callsToAssignCh:
				
				fmt.Printf("As: state: %v\n", elevators)
				assignments := assign(elevators, calls)
				//fmt.Printf("assigned:%v\n", assignments)
				assignmentsToSlaveCh <- assignments
				assignmentsToSlaveReceiver <- assignments
				fmt.Println("As:Succeded")
			default:
			}
		}
	}

}

func assign(elevators [config.N_ELEVATORS]slave.Elevator, callsToAssign AssignCalls) [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool { 

	hraExecutable := ""

	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}
	
	input := transformInput(elevators, callsToAssign) // transforms input from worldview to HRAInput

	fmt.Println("Input to assigner: ", string(input))

	// assign and returns output in json format
	outputJsonFormat, errAssign := exec.Command("heislab/Project-resources/cost_fns/hall_request_assigner/"+hraExecutable, "-i", string(input)).CombinedOutput()

	if errAssign != nil {
		fmt.Println("Error occured when assigning: ", errAssign,", output: ", string(outputJsonFormat))
	}

	// transforms output from json format to the correct ouputformat
	output := transformOutput(outputJsonFormat, callsToAssign)

	// make sure cab calls are not overwritten if elevator is stuck or not alive
	for elev := range(config.N_ELEVATORS) {
		for floor := range(config.N_FLOORS){
			output[elev][floor][2] = callsToAssign.Calls.CabCalls[elev][floor]
		}
	}

	fmt.Println("Output from assigner: ", output)

	return output
}

func transformInput(elevators [config.N_ELEVATORS]slave.Elevator, callsToAssign AssignCalls) []byte {

	input := HRAInput{
		HallRequests: callsToAssign.Calls.HallCalls[:],
		States:       map[string]HRAElevState{},
	}

	// adding all alive elevators to the input map
	for i := range(config.N_ELEVATORS) {
		if callsToAssign.AliveElevators[i] {
			input.States[strconv.Itoa(elevators[i].ID)] = HRAElevState{
				Floor:       elevators[i].Floor,
				Behavior:    behaviorMap[elevators[i].Behaviour],
				Direction:   directionMap[elevators[i].Direction],
				CabRequests: callsToAssign.Calls.CabCalls[i][:],
			}
		}
	}

	inputJsonFormat, errMarsial := json.Marshal(input)

	if errMarsial != nil {
		fmt.Println("Error using json.Marshal: ", errMarsial)
	}

	return inputJsonFormat
}

func transformOutput(outputJsonFormat []byte, callsToAssign AssignCalls) [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool {
	output := [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool{}
	tempOutput := new(map[string][config.N_FLOORS][2]bool)

	errUnmarshal := json.Unmarshal(outputJsonFormat, &tempOutput)

	if errUnmarshal != nil {
		fmt.Println("Error using json.Unmarshal: ", errUnmarshal)
	}

	for elevatorKey, tempElevatorOrders := range *tempOutput {
		elevatorId, err_convert := strconv.Atoi(elevatorKey)

		elevatorOrders := [config.N_FLOORS][config.N_BUTTONS]bool{}

		if err_convert != nil {
			fmt.Println("Error occured when converting to right assign format: ", err_convert)
		}

		for floor := range config.N_FLOORS {
			// appending cab calls from worldview of each floor to the output
			elevatorOrders[floor] = [3]bool{tempElevatorOrders[floor][0], tempElevatorOrders[floor][1], callsToAssign.Calls.CabCalls[elevatorId][floor]}
		}

		output[elevatorId] = elevatorOrders
	}

	return output
}

// TODO sjekke om det finnes elevators i live og som ikke er stuck - hvis ikke så assignes alt til heisen som er master (må ta inn own id) - er det nødvendig? 