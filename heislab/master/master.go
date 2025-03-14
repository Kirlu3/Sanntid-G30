package master

import (
	"fmt"
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func Master(initCalls Calls, Id int) {
	fmt.Println(Id, "entered master mode")

	callsUpdateCh := make(chan struct{Calls Calls; AddCall bool}, 2)
	callsToAssignCh := make(chan struct{Calls Calls; AliveElevators [config.N_ELEVATORS]bool})

	stateUpdateCh := make(chan slave.Elevator)
	assignmentsToSlaveCh := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	assignmentsToSlaveReceiverCh := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool, 2)

	masterTxEnable := make(chan bool)

	go peers.Transmitter(config.MasterUpdatePort, strconv.Itoa(Id), masterTxEnable)

	go backupAckRx(callsUpdateCh, callsToAssignCh, initCalls, Id)
	go assignOrders(stateUpdateCh, callsToAssignCh, assignmentsToSlaveCh, assignmentsToSlaveReceiverCh)

	go receiveMessagesFromSlaves(stateUpdateCh, callsUpdateCh, assignmentsToSlaveReceiverCh)
	go sendMessagesToSlaves(assignmentsToSlaveCh)

	// the program crashes and restarts when it should go to backup mode
	select {}
}
