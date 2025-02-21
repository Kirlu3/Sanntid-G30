package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
)

type Placeholder int

func Master(initWorldview slave.WorldView, masterUpdateCh chan peers.PeerUpdate, masterTxEnable chan bool, masterWorldViewTx chan slave.WorldView,
	masterWorldViewRx chan slave.WorldView, backupWorldViewRx chan slave.WorldView) {
	masterTxEnable <- true
	fmt.Println(initWorldview.OwnId, " entered master mode")

	// CHANNELS THAT GO ROUTINES WILL COMMUNICATE ON
	requestAssignment := make(chan struct{})     // currently none -> stateManager  | if anyone writes to this channel orders are reassigned, superfluous
	slaveUpdate := make(chan slave.EventMessage) // receiveMessagesFromSlaves -> stateManager
	backupUpdate := make(chan []string)          // trackAliveBackups -> stateManager
	mergeState := make(chan slave.WorldView)     // lookForOtherMasters -> stateManager
	stateToBackup := make(chan slave.WorldView)  // stateManager -> sendStateToBackups
	aliveBackups := make(chan []string)          // stateManager -> receiveBackupAck
	requestBackupAck := make(chan slave.Calls)   // stateManager -> receiveBackupAck
	stateToAssign := make(chan slave.WorldView)  // stateManager -> assignOrders
	orderAssignments := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)      // assignOrders -> sendMessagesToSlaves | [][]int wont work, need [][][]int or struct or something
	lightsToSlave := make(chan slave.Calls)      // receiveBackupAck -> sendMessagesToSlaves
	endMasterPhase := make(chan struct{})        // lookForOtherMasters -> Master | when a master with higher pri is found we end the master phase by writing to this channel

	// EXAMPLE OF POSSIBLE THREADS IN MASTER
	// go establishConnectionsToSlaves() // i have no idea how this is done or if this go routine makes sense

	// change in state happens when: message is received from slave, alive signal is lost from backup, mergeMaster
	go receiveMessagesFromSlaves(slaveUpdate)
	go stateManager(initWorldview, requestAssignment, slaveUpdate, backupUpdate, mergeState, stateToBackup, aliveBackups, requestBackupAck, stateToAssign)
	go sendStateToBackups(stateToBackup, masterWorldViewTx, initWorldview)
	go trackAliveBackups(backupUpdate)
	go receiveBackupAck(requestBackupAck, aliveBackups, lightsToSlave, backupWorldViewRx)
	go assignOrders(stateToAssign, orderAssignments)         // IMPORTANT: is it ok to assign an unconfirmed order? i think yes
	go sendMessagesToSlaves(orderAssignments, lightsToSlave) // orders (+ lights?)
	go lookForOtherMasters(endMasterPhase, masterUpdateCh, initWorldview.OwnId)

	<-endMasterPhase
	masterTxEnable <- false
	// to end the goroutines, close their channels (and add logic in the goroutines to return when channels are closed)
	// end all goroutines and return (to backup state) (if we learn that there are other masters with higher priority?)
	// does this master/backups structure make sense?

}
