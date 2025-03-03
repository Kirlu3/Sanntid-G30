package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type Placeholder int

func Master(
	initCalls BackupCalls,
	masterCallsTx chan<- BackupCalls,
	masterCallsRx <-chan BackupCalls,
	backupCallsRx <-chan BackupCalls,
	masterTxEnable chan<- bool,
	backupsUpdateCh <-chan peers.PeerUpdate,
) {
	masterTxEnable <- true
	fmt.Println(initCalls.Id, "entered master mode")

	// PLANNED NEW GO ROUTINES, NOTE THAT backupAckRx spawns some new go routines
	// go slaveStateRx()
	// go slaveCallsRx()

	callsUpdateCh := make(chan UpdateCalls, 2)
	callsToAssignCh := make(chan AssignCalls)

	stateUpdateCh := make(chan slave.Elevator)
	assignmentsToSlaveCh := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	assignmentsToSlaveReceiver := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool, 2)

	go backupAckRx(callsUpdateCh, callsToAssignCh, initCalls, masterCallsTx, masterCallsRx, backupCallsRx, backupsUpdateCh)
	go assignOrders(stateUpdateCh, callsToAssignCh, assignmentsToSlaveCh, assignmentsToSlaveReceiver)

	go receiveMessagesFromSlaves(stateUpdateCh, callsUpdateCh, assignmentsToSlaveReceiver) //starts other go routines
	go sendMessagesToSlaves(assignmentsToSlaveCh)                                          // orders (+ lights?) ??

	select {}
	// to end the goroutines, close their channels (and add logic in the goroutines to return when channels are closed)
	// end all goroutines and return (to backup state) (if we learn that there are other masters with higher priority?)
	// does this master/backups structure make sense?

}

//Comment: changed so assigned orders go through the state manager to update the states of the slaves, this should simply allow the slaves to clear orders immediately
