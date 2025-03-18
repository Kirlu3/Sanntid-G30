package master

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"slices"
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
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

/*
slaveStateUpdateChan receives updates about the state of the elevators

callsToAssignChan receives the calls that should be assigned and a list over the alive elevators

callsToSlaveChan sends the assigned orders to the function that handles sending them to the slaves
*/
func assignCalls( //slaveStateUpdateChan, callsToAssignChan, callsToSlaveChan
	slaveStateUpdateChan <-chan slave.Elevator,
	callsToAssignChan <-chan struct {
		Calls          Calls
		AliveElevators [config.N_ELEVATORS]bool
	},
	callsToSlaveChan chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	id int,
) {

	elevators := [config.N_ELEVATORS]slave.Elevator{} // consider waiting for state init
	calls := AssignCalls{}

	for i := range config.N_ELEVATORS {
		elevators[i].ID = i // suggested fix to assigner init bug
	}

	for {
		select {
		case stateUpdate := <-slaveStateUpdateChan:
			elevators[stateUpdate.ID] = stateUpdate
			fmt.Println("As:Received new states")

		case calls = <-callsToAssignChan:
			fmt.Printf("As: state: %v\n", elevators)
		}
		var assignedCalls [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
		for i, elevator := range elevators {
			if elevator.Stuck {
				calls.AliveElevators[i] = false // acts as if the elevator is dead if it is stuck
			}
		}
		if !slices.Contains(calls.AliveElevators[:], true) {
			calls.AliveElevators[id] = true
		}
		assignedCalls = assign(elevators, calls)
		callsToSlaveChan <- assignedCalls
		fmt.Println("As:Succeded")
	}
}

func assignAllCallsToMaster(calls Calls, id int) [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool {
	var assignedCalls [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	for floor := range config.N_FLOORS {
		assignedCalls[id][floor][elevio.BT_HallDown] = calls.HallCalls[floor][elevio.BT_HallDown]
		assignedCalls[id][floor][elevio.BT_HallUp] = calls.HallCalls[floor][elevio.BT_HallUp]
	}
	for elevator := range config.N_ELEVATORS {
		for floor := range config.N_FLOORS {
			assignedCalls[elevator][floor][elevio.BT_Cab] = calls.CabCalls[elevator][floor]
		}
	}
	return assignedCalls
}

/*
Input: the masters view of the elevator states and the calls that should be assigned

Output: an array containing what calls go to which elevator
*/
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

	input := transformInput(elevators, callsToAssign)

	fmt.Println("Input to assigner: ", string(input))

	// assigns and returns output in json format
	outputJsonFormat, errAssign := exec.Command("heislab/"+hraExecutable, "-i", string(input)).CombinedOutput()

	if errAssign != nil {
		fmt.Println("Error occured when assigning: ", errAssign, ", output: ", string(outputJsonFormat))
	}

	// transforms output from json format to the correct ouputformat
	output := transformOutput(outputJsonFormat, callsToAssign)

	// make sure cab calls are not overwritten if elevator is stuck or not alive
	for elev := range config.N_ELEVATORS {
		for floor := range config.N_FLOORS {
			output[elev][floor][elevio.BT_Cab] = callsToAssign.Calls.CabCalls[elev][floor]
		}
	}

	fmt.Println("Output from assigner: ", output)

	return output
}

/*
Input: the state of the elevators and the calls that should be assigned

Output: JSON encoding of the input
*/
func transformInput(elevators [config.N_ELEVATORS]slave.Elevator, callsToAssign AssignCalls) []byte {

	input := HRAInput{
		HallRequests: callsToAssign.Calls.HallCalls[:],
		States:       map[string]HRAElevState{},
	}

	// adding all alive elevators to the input map
	for i := range config.N_ELEVATORS {
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

/*
Input: JOSN encoding of the assigned calls

Output: an array of the assigned calls
*/
func transformOutput(outputJsonFormat []byte, callsToAssign AssignCalls) [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool {
	output := [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool{}
	tempOutput := new(map[string][config.N_FLOORS][config.N_BUTTONS - 1]bool)

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
			elevatorOrders[floor] = [3]bool{tempElevatorOrders[floor][elevio.BT_HallUp], tempElevatorOrders[floor][elevio.BT_HallDown], callsToAssign.Calls.CabCalls[elevatorId][floor]}
		}

		output[elevatorId] = elevatorOrders
	}

	return output
}
