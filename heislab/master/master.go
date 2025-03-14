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
	masterToSlaveOfflineCh chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	slaveToMasterOfflineButton <-chan slave.ButtonMessage,
	slaveToMasterOfflineElevator <-chan slave.Elevator,
) {
	fmt.Println(Id, "entered master mode")

	callsUpdateCh := make(chan struct {
		Calls   Calls
		AddCall bool
	}, 2)
	callsToAssignCh := make(chan struct {
		Calls          Calls
		AliveElevators [config.N_ELEVATORS]bool
	})

	stateUpdateCh := make(chan slave.Elevator)
	assignmentsToSlaveCh := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	masterTxEnable := make(chan bool)

	go peers.Transmitter(config.MasterUpdatePort, strconv.Itoa(Id), masterTxEnable)

	go backupAckRx(callsUpdateCh, callsToAssignCh, initCalls, Id)
	go assignOrders(stateUpdateCh, callsToAssignCh, assignmentsToSlaveCh)

	go receiveButtonPress(callsUpdateCh, slaveToMasterOfflineButton)
	go receiveElevatorUpdate(stateUpdateCh, callsUpdateCh, slaveToMasterOfflineElevator)
	go sendMessagesToSlaves(assignmentsToSlaveCh, masterToSlaveOfflineCh)

	// the program is crashed and restarted when it should go back to backup mode
	select {}
}
