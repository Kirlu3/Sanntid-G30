package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type Placeholder int

// func Master(initWorldview slave.WorldView, masterUpdateCh chan peers.PeerUpdate, masterTxEnable chan bool, masterWorldViewTx chan slave.WorldView,
// masterWorldViewRx chan slave.WorldView, backupWorldViewRx chan slave.WorldView, backupsUpdateCh chan peers.PeerUpdate) {
func Master(
	initCalls slave.BackupCalls,
	masterCallsTx chan<- slave.BackupCalls,
	masterCallsRx <-chan slave.BackupCalls,
	backupCallsRx <-chan slave.BackupCalls,
	masterTxEnable chan<- bool,
	backupsUpdateCh <-chan peers.PeerUpdate,
) {
	masterTxEnable <- true
	fmt.Println(initCalls.Id, " entered master mode")

	// PLANNED NEW GO ROUTINES, NOTE THAT backupAckRx spawns some new go routines
	// go slaveStateRx()
	// go slaveCallsRx()

	callsUpdateCh := make(chan slave.Calls)
	callsToAssignCh := make(chan slave.AssignCalls)
	endMasterPhaseCh := make(chan struct{})

	stateUpdateCh := make(chan [config.N_ELEVATORS]slave.Elevator)
	assignmentsToSlaveCh := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)

	go backupAckRx(callsUpdateCh, callsToAssignCh, endMasterPhaseCh, initCalls, masterCallsTx, masterCallsRx, backupCallsRx, backupsUpdateCh)
	go assignOrders(stateUpdateCh, callsToAssignCh, assignmentsToSlaveCh)

	go receiveMessagesFromSlaves(stateUpdateCh, callsUpdateCh) //starts other go routines
	go sendMessagesToSlaves(assignmentsToSlaveCh)              // orders (+ lights?) ??

	<-endMasterPhaseCh
	masterTxEnable <- false
	// to end the goroutines, close their channels (and add logic in the goroutines to return when channels are closed)
	// end all goroutines and return (to backup state) (if we learn that there are other masters with higher priority?)
	// does this master/backups structure make sense?

}

//Comment: changed so assigned orders go through the state manager to update the states of the slaves, this should simply allow the slaves to clear orders immediately
