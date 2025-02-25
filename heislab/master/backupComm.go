package master

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
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

func aliveBackupsRx(aliveBackupsCh chan<- []string, backupsUpdateCh chan peers.PeerUpdate) {
	a := <-backupsUpdateCh
	var aliveBackups []string = a.Peers
	aliveBackupsCh <- aliveBackups
	for {
		a := <-backupsUpdateCh
		fmt.Printf("Backups update:\n")
		fmt.Printf("  Backups:    %q\n", a.Peers)
		fmt.Printf("  New:        %q\n", a.New)
		fmt.Printf("  Lost:       %q\n", a.Lost)
		aliveBackups = a.Peers
		if len(a.Lost) != 0 || a.New != "" {
			aliveBackupsCh <- aliveBackups
		}
	}
}

// when all aliveBackups have the same calls as requestBackupAck send lightsToSlave
func receiveBackupAck(OwnId string, requestBackupAckCh <-chan slave.Calls, aliveBackupsCh <-chan []string, aliveBackupsToManagerCh chan<- []string, callsToAssign chan<- slave.Calls, backupWorldViewRx <-chan slave.WorldView) {
	ID, _ := strconv.Atoi(OwnId)
	var aliveBackups []string = <- aliveBackupsCh
	var acksReceived [config.N_ELEVATORS]bool
	var calls slave.Calls
	newCalls := false
mainLoop:
	for {
		select {
		case calls = <-requestBackupAckCh: // when we receive new calls reset all acks
			newCalls = true
			for i := range acksReceived {
				acksReceived[i] = false
			}
			acksReceived[ID] = true
		default:
		}

		select {
		case a := <-backupWorldViewRx: // set ack for backup if it has the same calls
			if (sameCalls(slave.Calls{HallCalls: a.HallCalls, CabCalls: a.CabCalls}, calls)) {
				i, _ := strconv.Atoi(a.OwnId)
				acksReceived[i] = true
			}
		default:
		}

		select {
		case aliveBackups = <-aliveBackupsCh:
			aliveBackupsToManagerCh <- aliveBackups
		default:
		}

		for _, backup := range aliveBackups { // if some alive backups havent given ack, continue main loop
			i, _ := strconv.Atoi(backup)
			if !acksReceived[i] {
				continue mainLoop
			}
		}
		if newCalls {
			fmt.Println("BC: Sending calls")
			callsToAssign <- calls // orders to assign
			newCalls = false
		}
	}
}

// returns true if two Calls structs are the same, might not be necessary, i think go can compare the structs directly with calls1 == calls2, look into this
func sameCalls(calls1 slave.Calls, calls2 slave.Calls) bool {
	for i := 0; i < config.N_ELEVATORS; i++ {
		for j := 0; j < config.N_FLOORS; j++ {
			if calls1.CabCalls[i][j] != calls2.CabCalls[i][j] {
				return false
			}
		}
	}
	for i := 0; i < config.N_FLOORS; i++ {
		for j := 0; j < config.N_BUTTONS-1; j++ {
			if calls1.HallCalls[i][j] != calls2.HallCalls[i][j] {
				return false
			}
		}
	}
	return true
}
