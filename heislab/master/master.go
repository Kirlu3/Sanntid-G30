package master

import (
	"fmt"

	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
	"github.com/mohae/deepcopy"
)

type Placeholder int


// the master can make BIG decisions (like turning on lights)
// if it has received conformation that all other alive backups mirror its own state (but does this make more sense with tcp?)
// if no message from x for some time: mark x as dead
// reassign orders assigned to x (SMALL decision, can be done anytime?)

// i am somewhat concerned about slices in the worldview struct being sent by reference, consider doing something about this
func Master(initWorldview Slave.WorldView) {
	fmt.Println(initWorldview.OwnId, " entered master mode")
	// CHANNELS THAT GO ROUTINES WILL COMMUNICATE ON
	doAssignOrders := make(chan struct{}) // from

	orderAssignments := make(chan [][]int) // [][]int wont work, need [][][]int or struct or something

	endMasterPhase := make(chan struct{})

	// EXAMPLE OF POSSIBLE THREADS IN MASTER
	// go establishConnectionsToSlaves()

	// change in state happens when: message is received from slave, alive signal is lost from backup, mergeMaster
	go receiveMessagesFromSlaves(slaveUpdate) // all relevant info from slave
	// this is a lot, everything that reads/writes the state goes through the statemanager
	go stateManager(initWorldview, requestAssignment, slaveUpdate, backupUpdate, mergeState, stateToBackup, aliveBackups, stateToAssign)
	go sendStateToBackups(stateToBackup)
	go trackAliveBackups(backupUpdate)                      // these should perhaps be one function?
	go receiveBackupConformation(aliveBackups, updateLight) // these should perhaps be one function?
	go turnOnLight(updateLight, requestAssignment)          // requestAssignment is only necessary if we want to try to turn on the lights BEFORE assigning the order
	// however it is still possible for the order to be assigned in the meantime so maybe just dont ever bother?
	go turnOffLight()
	go assignOrders(stateToAssign, orderAssignments) // IMPORTANT: is it ok to assign an unconfirmed order? i think yes
	go sendMessagesToSlaves(orderAssignments)        // orders (+ lights?)
	go lookForOtherMasters(endMasterPhase)

	<-endMasterPhase
	// to end the goroutines, close their channels (and add logic in the goroutines to return when channels are closed)
	// end all goroutines and return (to backup state) (if we learn that there are other masters with higher priority?)
	// does this master/backups structure make sense?

}

func receiveMessagesFromSlaves(slaveUpdate chan Slave.EventMessage) {
	// receive a message from slave and put it on slaveUpdate
	var message Slave.EventMessage
	slaveUpdate <- message
}


func stateManager(initWorldview Slave.WorldView, requestAssignment chan struct{}, slaveUpdate chan Slave.EventMessage, backupUpdate []string, mergeState chan Slave.WorldView,
	stateToBackup chan Slave.WorldView, aliveBackups []string, stateToAssign chan Slave.WorldView) {
	worldview := deepcopy.Copy(initWorldview).(Slave.WorldView)
	for {
		select {
		case <-requestAssignment:
			stateToAssign <- worldview

		case a := <-slaveUpdate:
			if (a.Event == Slave.Button) {
				worldview.
				// update the button in question 
			}
		
		}
	}
}

func lookForOtherMasters(endMasterPhase chan struct{}) {
	// listen on the master alive channel to see if there are any other masters

	return
}

func turnOnLight() {
	// is turning on lights different from turning off lights?

	// to turn on light we need:
}

func assignOrders(stateToAssign chan Slave.WorldView, orderAssignments chan [][]int) {
	for {
		select {
		case state :=<-stateToAssign: // i think it would be better to instead do case <-state since assignments should generally be made when state is updated
			orderAssignments <- costfunc(state) // the D cost algorithm
		}
	}
}

func sendMessagesToSlaves(orderAssignments chan [][]int) {
	for {
		select {
		case <-orderAssignments:
			// for each elevator: send the orders that it has been assigned
		}
	}
}
