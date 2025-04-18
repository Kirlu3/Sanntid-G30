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
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

var behaviorMap = map[slave.ElevatorBehaviour]string{
	slave.BehaviourIdle:     "idle",
	slave.BehaviourMoving:   "moving",
	slave.BehaviourDoorOpen: "doorOpen",
}

var directionMap = map[slave.ElevatorDirection]string{
	slave.DirectionDown: "down",
	slave.DirectionStop: "stop",
	slave.DirectionUp:   "up",
}

/*
# The assigner runs in this goroutine. It collects all necessary information to assign calls to the elevators, assigns them and sends the assigned calls to the slaveComm module.

Input: slaveStateUpdateChan, callsToAssignChan, callsToSlaveChan, masterId

slaveStateUpdateChan: receives updates about the state of the elevators

callsToAssignChan: receives the calls that should be assigned and a list over the alive elevators

callsToSlaveChan: sends the assigned calls to a routine that sends them to the slaves
*/
func assigner(
	slaveStateUpdateChan <-chan slave.Elevator,
	callsToAssignChan <-chan struct {
		Calls          Calls
		AliveElevators [config.NumElevators]bool
	},
	callsToSlaveChan chan<- [config.NumElevators][config.NumFloors][config.NumBtns]bool,
	ownId int,
) {

	elevators := [config.NumElevators]slave.Elevator{}
	callsToAssignUpdate := <-callsToAssignChan
	calls := callsToAssignUpdate.Calls
	aliveElevators := callsToAssignUpdate.AliveElevators

	for i := range config.NumElevators {
		elevators[i].ID = i
	}

	for {
		select {
		case stateUpdate := <-slaveStateUpdateChan:
			elevators[stateUpdate.ID] = stateUpdate

		case callsToAssignUpdate = <-callsToAssignChan:
			calls = callsToAssignUpdate.Calls
			aliveElevators = callsToAssignUpdate.AliveElevators
		}

		availableElevators := aliveAndNotStuck(aliveElevators, elevators)

		if !slices.Contains(availableElevators[:], true) {
			availableElevators[ownId] = true
		}
		assignedCalls := assignCalls(elevators, calls, availableElevators)
		callsToSlaveChan <- assignedCalls
	}
}

/*
Input: The alive elevators and their elevator states

Returns: an array containing the alive elevators that are not stuck
*/
func aliveAndNotStuck(aliveElevators [3]bool, elevators [3]slave.Elevator) [config.NumElevators]bool {
	var availableElevators [config.NumElevators]bool
	for elev := range config.NumElevators {
		availableElevators[elev] = aliveElevators[elev] && !elevators[elev].Stuck
	}
	return availableElevators
}

/*
# This function assigns calls to the elevators using the hall request assigner.

Input: the masters view of the elevator states and the calls that should be assigned

Output: an array containing what calls go to which elevator
*/
func assignCalls(
	elevators [config.NumElevators]slave.Elevator,
	calls Calls,
	availableElevators [config.NumElevators]bool,
) [config.NumElevators][config.NumFloors][config.NumBtns]bool {
	hraExecutable := ""

	switch runtime.GOOS {
	case "linux":
		hraExecutable = "hall_request_assigner"
	case "windows":
		hraExecutable = "hall_request_assigner.exe"
	default:
		panic("OS not supported")
	}

	input := transformInputToJSON(elevators, calls, availableElevators)
	outputJsonFormat, errAssign := exec.Command("heislab/"+hraExecutable, "-i", string(input)).CombinedOutput()

	if errAssign != nil {
		fmt.Println("Error occured when assigning: ", errAssign, ", output: ", string(outputJsonFormat))
	}

	output := transformOutputFromJSON(outputJsonFormat, calls)

	for elev := range config.NumElevators {
		for floor := range config.NumFloors {
			output[elev][floor][elevio.BT_Cab] = calls.CabCalls[elev][floor]
		}
	}

	return output
}

/*
Input: the state of the elevators and the calls that should be assigned

Output: JSON encoding of the input
*/
func transformInputToJSON(elevators [config.NumElevators]slave.Elevator, calls Calls, availableElevators [config.NumElevators]bool) []byte {

	input := HRAInput{
		HallRequests: calls.HallCalls[:],
		States:       map[string]HRAElevState{},
	}

	for i := range config.NumElevators {
		if availableElevators[i] {
			input.States[strconv.Itoa(elevators[i].ID)] = HRAElevState{
				Floor:       elevators[i].Floor,
				Behavior:    behaviorMap[elevators[i].Behaviour],
				Direction:   directionMap[elevators[i].Direction],
				CabRequests: calls.CabCalls[i][:],
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
func transformOutputFromJSON(outputJsonFormat []byte, calls Calls) [config.NumElevators][config.NumFloors][config.NumBtns]bool {
	output := [config.NumElevators][config.NumFloors][config.NumBtns]bool{}
	tempOutput := new(map[string][config.NumFloors][config.NumBtns - 1]bool)

	errUnmarshal := json.Unmarshal(outputJsonFormat, &tempOutput)

	if errUnmarshal != nil {
		fmt.Println("Error using json.Unmarshal: ", errUnmarshal)
	}

	for elevatorKey, tempElevatorOrders := range *tempOutput {
		elevatorId, err_convert := strconv.Atoi(elevatorKey)

		elevatorOrders := [config.NumFloors][config.NumBtns]bool{}

		if err_convert != nil {
			fmt.Println("Error occured when converting to right assign format: ", err_convert)
		}

		for floor := range config.NumFloors {
			elevatorOrders[floor] = [3]bool{tempElevatorOrders[floor][elevio.BT_HallUp], tempElevatorOrders[floor][elevio.BT_HallDown], calls.CabCalls[elevatorId][floor]}
		}

		output[elevatorId] = elevatorOrders
	}

	return output
}
