package master

import (
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
	"github.com/mohae/deepcopy"
)

// consider running this as a nested function inside statemanager instead
func sendStateToBackups(stateToBackup chan slave.WorldView, masterWorldViewTx chan slave.WorldView, initWorldview slave.WorldView) {
	worldview := deepcopy.Copy(initWorldview).(slave.WorldView)
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
func receiveBackupAck(requestBackupAck chan slave.Calls, aliveBackups chan []string, lightsToSlave chan slave.Calls, backupWorldViewRx chan slave.WorldView) {
	// TODO: when aliveBackups gets a new message do:
}

func trackAliveBackups(backupUpdate chan []string) {
	// if we lose a node is easy: just tell stateManager which reassigns orders
	// gaining a node might be more complicated: the new node might require some additional information, e.g. which lights to set
}
