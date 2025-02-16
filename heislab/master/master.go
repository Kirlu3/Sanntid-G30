package master

import (
	"fmt"
	"slices"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
	"github.com/mohae/deepcopy"
)

type Placeholder int

func Master(initWorldview Slave.WorldView, masterUpdateCh chan peers.PeerUpdate, masterTxEnable chan bool, masterWorldViewTx chan Slave.WorldView,
	masterWorldViewRx chan Slave.WorldView, backupWorldViewRx chan Slave.WorldView) {
	masterTxEnable <- true
	fmt.Println(initWorldview.OwnId, " entered master mode")

	// CHANNELS THAT GO ROUTINES WILL COMMUNICATE ON
	requestAssignment := make(chan struct{})     // currently none -> stateManager  | if anyone writes to this channel orders are reassigned, superfluous
	slaveUpdate := make(chan Slave.EventMessage) // receiveMessagesFromSlaves -> stateManager
	backupUpdate := make(chan []string)          // trackAliveBackups -> stateManager
	mergeState := make(chan Slave.WorldView)     // lookForOtherMasters -> stateManager
	stateToBackup := make(chan Slave.WorldView)  // stateManager -> sendStateToBackups
	aliveBackups := make(chan []string)          // stateManager -> receiveBackupAck
	requestBackupAck := make(chan Slave.Calls)   // stateManager -> receiveBackupAck
	stateToAssign := make(chan Slave.WorldView)  // stateManager -> assignOrders
	orderAssignments := make(chan [][]int)       // assignOrders -> sendMessagesToSlaves | [][]int wont work, need [][][]int or struct or something
	lightsToSlave := make(chan Slave.Calls)      // receiveBackupAck -> sendMessagesToSlaves
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

func receiveMessagesFromSlaves(slaveUpdate chan Slave.EventMessage) {
	// receive a message from slave and put it on slaveUpdate
	var message Slave.EventMessage // TODO
	slaveUpdate <- message
}

// it is important that this function doesnt block
func stateManager(initWorldview Slave.WorldView, requestAssignment chan struct{}, slaveUpdate chan Slave.EventMessage, backupUpdate chan []string,
	mergeState chan Slave.WorldView, stateToBackup chan Slave.WorldView, aliveBackups chan []string, requestBackupAck chan Slave.Calls,
	stateToAssign chan Slave.WorldView) {
	// aliveBackups might be redundant
	worldview := deepcopy.Copy(initWorldview).(Slave.WorldView)
	for {
		select {
		case <-requestAssignment:
			stateToAssign <- worldview

		case slaveMessage := <-slaveUpdate:
			slaveId := int(slaveMessage.Elevator.Id[0] - '0')
			switch slaveMessage.Event {

			case Slave.Button:
				if slaveMessage.Btn.Button == elevio.BT_Cab {
					worldview.CabCalls[slaveId][slaveMessage.Btn.Floor][slaveMessage.Btn.Button] = slaveMessage.Check
				} else {
					worldview.HallCalls[slaveMessage.Btn.Floor] = slaveMessage.Check // do we have to be careful when removing order? i dont think so
				}
				stateToAssign <- deepcopy.Copy(worldview).(Slave.WorldView)
				requestBackupAck <- Slave.Calls{
					HallCalls: deepcopy.Copy(worldview.HallCalls).([config.N_FLOORS]bool),
					CabCalls:  deepcopy.Copy(worldview.HallCalls).([10][config.N_FLOORS][2]bool),
				}
				break

			case Slave.FloorArrival:
				worldview.Elevators[slaveId] = slaveMessage.Elevator // i think it makes sense to update the whole state, again consider deepcopy
				// should we reassign orders here?
				break

			case Slave.Stuck:
				worldview.Elevators[slaveId].Stuck = slaveMessage.Check
				stateToAssign <- deepcopy.Copy(worldview).(Slave.WorldView)
				break

			default:
				panic("invalid message event from slave")
			}

		case backups := <-backupUpdate:
			for i := range worldview.AliveElevators {
				worldview.AliveElevators[i] = false
			}
			for _, aliveIdx := range backups {
				worldview.AliveElevators[int(aliveIdx[0]-'0')] = true
			}
			stateToAssign <- deepcopy.Copy(worldview).(Slave.WorldView)
			// maybe forward the update to receiveBackupAck on aliveBackups channel

		case otherMasterState := <-mergeState:
			fmt.Printf("otherMasterState: %v\n", otherMasterState)
			// inherit calls from otherMaster TODO
			stateToAssign <- deepcopy.Copy(worldview).(Slave.WorldView)

		}
		stateToBackup <- deepcopy.Copy(worldview).(Slave.WorldView)

	}
}

// consider using the masterWorldViewRx instead
func lookForOtherMasters(endMasterPhase chan struct{}, masterUpdateCh chan peers.PeerUpdate, ownId string) {
	// lookForMastersLoop:
	for {
		select {
		case masterUpdate := <-masterUpdateCh:
			if slices.Min(masterUpdate.Peers) < ownId {
				// end the master phase, but the other master needs to get our state first!
			}
			if slices.Max(masterUpdate.Peers) > ownId {
				// make sure the stateManager gets the mergeState message
			}
		}
	}

}

func assignOrders(stateToAssign chan Slave.WorldView, orderAssignments chan [][]int) {
	for {
		select {
		case state := <-stateToAssign:
			fmt.Printf("state: %v\n", state)
			var assignments [][]int // TODO: call the D cost algo
			orderAssignments <- assignments
		}
	}
}

func sendMessagesToSlaves(orderAssignments chan [][]int, lightsToSlave chan Slave.Calls) {
	for {
		select {
		case a := <-orderAssignments:
			fmt.Printf("a: %v\n", a)
			// for each elevator: send the orders that it has been assigned
		case s := <-lightsToSlave:
			fmt.Printf("s: %v\n", s)
			// send the lights to all the elevators (all lights or only updates? i dont think we have a good way of checking diff because who actually knows what lights are on?
			// so maybe just send all lights for each elevator)
		}

	}
}

// consider running this as a nested function inside statemanager instead
func sendStateToBackups(stateToBackup chan Slave.WorldView, masterWorldViewTx chan Slave.WorldView, initWorldview Slave.WorldView) {
	worldview := deepcopy.Copy(initWorldview).(Slave.WorldView)
	for {
		select {
		case worldview = <-stateToBackup:
			masterWorldViewTx <- worldview
		case <-time.After(config.MasterMessagePeriodSeconds):
			masterWorldViewTx <- worldview
		}
	}

}

// when all aliveBackups have the same calls as requestBackupAck send lightsToSlave
func receiveBackupAck(requestBackupAck chan Slave.Calls, aliveBackups chan []string, lightsToSlave chan Slave.Calls, backupWorldViewRx chan Slave.WorldView) {
	// TODO: when aliveBackups gets a new message do:
}

func trackAliveBackups(backupUpdate chan []string) {
	// if we lose a node is easy: just tell stateManager which reassigns orders
	// gaining a node might be more complicated: the new node might require some additional information, e.g. which lights to set
}
