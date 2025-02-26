package master

import (
	"fmt"
	"strconv"

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
	// go slaveCallsTx()

	go backupAckRx(callsUpdateCh, callsToAssignCh, initCalls, masterCallsTx, masterCallsRx, backupCallsRx, backupsUpdateCh)
	go assignOrders(stateUpdateCh, callsToAssignCh, assignmentsToSlaveCh)

	// CHANNELS THAT GO ROUTINES WILL COMMUNICATE ON
	requestAssignment := make(chan struct{})                                                   // currently none -> stateManager  | if anyone writes to this channel orders are reassigned, superfluous
	slaveUpdate := make(chan slave.EventMessage)                                               // receiveMessagesFromSlaves -> stateManager
	backupUpdate := make(chan []string)                                                        // trackAliveBackups -> stateManager
	mergeState := make(chan slave.BackupCalls)                                                 // lookForOtherMasters -> stateManager
	stateToBackup := make(chan slave.WorldView)                                                // stateManager -> sendStateToBackups
	aliveBackupsCh := make(chan []string)                                                      // stateManager -> receiveBackupAck
	aliveBackupsToManagerCh := make(chan []string)                                             // stateManager -> receiveBackupAck
	requestBackupAck := make(chan slave.Calls)                                                 // stateManager -> receiveBackupAck
	stateToAssign := make(chan slave.WorldView)                                                // stateManager -> assignOrders
	endMasterPhase := make(chan struct{})                                                      // lookForOtherMasters -> Master | when a master with higher pri is found we end the master phase by writing to this channel
	toSlaveCh := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)        //stateManager -> sendMessagesToSlaves
	assignedRequests := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool) //assignOrders -> stateManager
	callsToAssign := make(chan slave.AssignCalls)
	// EXAMPLE OF POSSIBLE THREADS IN MASTER
	// go establishConnectionsToSlaves() // i have no idea how this is done or if this go routine makes sense

	var initWorldview slave.WorldView
	initWorldview.CabCalls = initCalls.Calls.CabCalls
	initWorldview.HallCalls = initCalls.Calls.HallCalls
	// change in state happens when: message is received from slave, alive signal is lost from backup, mergeMaster
	receiveMessagesFromSlaves(slaveUpdate) //starts other go routines
	go stateManager(initWorldview, requestAssignment, slaveUpdate, backupUpdate, mergeState, stateToBackup, aliveBackupsToManagerCh, requestBackupAck, stateToAssign, assignedRequests, toSlaveCh, endMasterPhase)
	go backupsTx(stateToBackup, masterCallsTx, initCalls)
	go aliveBackupsRx(aliveBackupsCh, backupsUpdateCh)
	go receiveBackupAck(strconv.Itoa(initCalls.Id), requestBackupAck, aliveBackupsCh, aliveBackupsToManagerCh, callsToAssign, backupCallsRx)
	go assignOrders(stateToAssign, assignedRequests, callsToAssign) // IMPORTANT: is it ok to assign an unconfirmed order? i think yes
	go sendMessagesToSlaves(toSlaveCh)                              // orders (+ lights?) ??
	go lookForOtherMasters(masterCallsRx, initCalls.Id, mergeState)

	<-endMasterPhase
	masterTxEnable <- false
	// to end the goroutines, close their channels (and add logic in the goroutines to return when channels are closed)
	// end all goroutines and return (to backup state) (if we learn that there are other masters with higher priority?)
	// does this master/backups structure make sense?

}

//Comment: changed so assigned orders go through the state manager to update the states of the slaves, this should simply allow the slaves to clear orders immediately
