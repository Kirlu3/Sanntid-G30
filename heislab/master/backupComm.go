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

// when all aliveBackups have the same calls as requestBackupAck send lightsToSlave
func receiveBackupAck(requestBackupAckCh chan slave.Calls, aliveBackupsCh chan []string, lightsToSlave chan slave.Calls, backupWorldViewRx chan slave.WorldView,
	backupsUpdateCh chan peers.PeerUpdate) {
	var aliveBackups []string
	// updating the peers
	go func() {
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
	}()
	var acksReceived [config.N_ELEVATORS]bool
	var calls slave.Calls
mainLoop:
	for {
		select {
		case calls = <-requestBackupAckCh: // when we receive new calls reset all acks
			for i := range acksReceived {
				acksReceived[i] = false
			}
		default:
			select {
			case a := <-backupWorldViewRx: // set ack for backup if it has the same calls
				if (sameCalls(slave.Calls{HallCalls: a.HallCalls, CabCalls: a.CabCalls}, calls)) {
					i, _ := strconv.Atoi(a.OwnId)
					acksReceived[i] = true
				}
			default:
			}
		}
		for _, backup := range aliveBackups { // if all the alive backups have given acks send light message to slave
			i, _ := strconv.Atoi(backup)
			if acksReceived[i] == false {
				continue mainLoop
			}
		}
		lightsToSlave <- calls
	}
}

// returns true if two Calls structs are the same
func sameCalls(calls1 slave.Calls, calls2 slave.Calls) bool {
	for i := 0; i < config.N_ELEVATORS; i++ {
		for j := 0; j < config.N_FLOORS; j++ {
			if calls1.CabCalls[i][j] != calls1.CabCalls[i][j] {
				return false
			}
		}
	}
	for i := 0; i < config.N_FLOORS; i++ {
		for j := 0; j < config.N_BUTTONS-1; j++ {
			if calls1.HallCalls[i][j] != calls1.HallCalls[i][j] {
				return false
			}
		}
	}
	return true
}
