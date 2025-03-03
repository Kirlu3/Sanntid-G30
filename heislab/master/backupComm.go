package master

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
)

func backupsTx(callsToBackupsCh <-chan Calls, masterCallsTx chan<- BackupCalls, initCalls BackupCalls) {
	calls := initCalls
	for {
		select {
		case calls.Calls = <-callsToBackupsCh:
			masterCallsTx <- calls
		case <-time.After(config.MasterMessagePeriodSeconds):
			masterCallsTx <- calls
		}
	}

}

func aliveBackupsRx(aliveBackupsCh chan<- []string, backupsUpdateCh <-chan peers.PeerUpdate) {
	// a := <-backupsUpdateCh
	var aliveBackups []string
	// aliveBackupsCh <- aliveBackups
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
func backupAckRx(
	callsUpdateCh <-chan UpdateCalls, //the message we get from slaveRx to calculate updated calls are doesnt have to be this type, but with this + calls we should be able to calculate what the updated calls should be
	callsToAssignCh chan<- AssignCalls,
	initCalls BackupCalls,
	masterCallsTx chan<- BackupCalls,
	masterCallsRx <-chan BackupCalls,
	backupCallsRx <-chan BackupCalls,
	backupsUpdateCh <-chan peers.PeerUpdate,
) {
	Id := initCalls.Id

	otherMasterCallsCh := make(chan BackupCalls)
	aliveBackupsCh := make(chan []string)
	callsToBackupsCh := make(chan Calls)

	// for all channels consider if it is nicer to start them here or in master()
	go lookForOtherMasters(otherMasterCallsCh, Id, masterCallsRx)
	go aliveBackupsRx(aliveBackupsCh, backupsUpdateCh)
	go backupsTx(callsToBackupsCh, masterCallsTx, initCalls)

	var aliveBackups []string
	var acksReceived [config.N_ELEVATORS]bool
	calls := initCalls.Calls
	wantReassignment := false
mainLoop:
	for {
		// fmt.Println("blocking?")
		select {
		case callsUpdate := <-callsUpdateCh:
			if callsUpdate.AddCall {
				calls = union(calls, callsUpdate.Calls)
			} else {
				calls = removeCalls(calls, callsUpdate.Calls)
			}
			callsToBackupsCh <- calls
			wantReassignment = true
			for i := range acksReceived {
				acksReceived[i] = false
			}
			acksReceived[Id] = true
		default:
		}

		select {
		case a := <-backupCallsRx: // set ack for backup if it has the same calls
			if a.Calls == calls && !acksReceived[a.Id] {
				fmt.Println("new backup state from", a.Id)
				acksReceived[a.Id] = true
			}
		default:
		}

		select {
		case aliveBackups = <-aliveBackupsCh:
			wantReassignment = true
		default:
		}

		select {
		case otherMasterCalls := <-otherMasterCallsCh:
			if otherMasterCalls.Id < Id && isCallsSubset(calls, otherMasterCalls.Calls) {
				fmt.Println("find a better way to restart the program")
				os.Exit(42) // intentionally crashing, program restarts automatically when exiting with code 42
			} else if otherMasterCalls.Id > Id {
				calls = union(calls, otherMasterCalls.Calls)
				callsToBackupsCh <- calls
				wantReassignment = true
			} else {
				fmt.Println("couldn't end master phase: other master has not accepted our calls")
			}
		default:
		}

		for _, backup := range aliveBackups { // if some alive backups havent given ack, continue main loop
			i, _ := strconv.Atoi(backup)
			if !acksReceived[i] {
				continue mainLoop
			}
		}
		if wantReassignment {
			fmt.Println("BC: Sending calls")
			var AliveElevators [config.N_ELEVATORS]bool
			for _, elev := range aliveBackups {
				idx, err := strconv.Atoi(elev)
				if err != nil {
					panic("BC got weird aliveElev")
				}
				AliveElevators[idx] = true
			}
			AliveElevators[Id] = true
			callsToAssignCh <- AssignCalls{Calls: calls, AliveElevators: AliveElevators} // orders to assign, do we ever block here?
			wantReassignment = false
		}
	}
}

// returns true if calls1 is a subset of calls2
func isCallsSubset(calls1 Calls, calls2 Calls) bool {
	for i := range config.N_ELEVATORS {
		for j := range config.N_FLOORS {
			if calls1.CabCalls[i][j] && !calls2.CabCalls[i][j] {
				return false
			}
		}
	}
	for i := range config.N_FLOORS {
		for j := range config.N_BUTTONS - 1 {
			if calls1.HallCalls[i][j] && !calls2.HallCalls[i][j] {
				return false
			}
		}
	}
	return true
}

// returns the union of the calls in calls1 and calls2
func union(calls1 Calls, calls2 Calls) Calls {
	var unionCalls Calls
	for i := range config.N_ELEVATORS {
		for j := range config.N_FLOORS {
			unionCalls.CabCalls[i][j] = calls1.CabCalls[i][j] || calls2.CabCalls[i][j]
		}
	}
	for i := range config.N_FLOORS {
		for j := range config.N_BUTTONS - 1 {
			unionCalls.HallCalls[i][j] = calls1.HallCalls[i][j] || calls2.HallCalls[i][j]
		}
	}
	return unionCalls
}

// returns calls \ removedCalls, where \ is set difference
func removeCalls(calls Calls, removedCalls Calls) Calls {
	updatedCalls := calls

	for i := range config.N_ELEVATORS {
		for j := range config.N_FLOORS {
			updatedCalls.CabCalls[i][j] = calls.CabCalls[i][j] && !removedCalls.CabCalls[i][j]
		}
	}
	for i := range config.N_FLOORS {
		for j := range config.N_BUTTONS - 1 {
			updatedCalls.HallCalls[i][j] = calls.HallCalls[i][j] && !removedCalls.HallCalls[i][j]
		}
	}
	return updatedCalls
}
