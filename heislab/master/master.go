package master

import (
	"fmt"
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func Master( 
	initCalls Calls,
	Id int,
	masterToSlaveCalls_offlineChan chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	slaveToMasterBtn_offlineChan <-chan slave.ButtonMessage,
	slaveToMasterElevState_offlineChan <-chan slave.Elevator,
) {
	fmt.Println(Id, "entered master mode")

	callsUpdateChan := make(chan struct {
		Calls   Calls
		AddCall bool
	}, 2)

	callsToAssignChan := make(chan struct {
		Calls          Calls
		AliveElevators [config.N_ELEVATORS]bool
	})

	stateUpdateChan := make(chan slave.Elevator)
	assignedCallsToSlaveChan := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	masterTransmitterEnableChan := make(chan bool)

	go peers.Transmitter(config.MasterUpdatePort, strconv.Itoa(Id), masterTransmitterEnableChan)

	go backupAckReceiver(callsUpdateChan, callsToAssignChan, initCalls, Id)
	go assignCalls(stateUpdateChan, callsToAssignChan, assignedCallsToSlaveChan)

	go receiveButtonPress(callsUpdateChan, slaveToMasterBtn_offlineChan)
	go receiveElevatorUpdate(stateUpdateChan, callsUpdateChan, slaveToMasterElevState_offlineChan)
	go sendMessagesToSlaves(assignedCallsToSlaveChan, masterToSlaveCalls_offlineChan)

	// the program is crashed and restarted when it should go back to backup mode
	select {}
}
