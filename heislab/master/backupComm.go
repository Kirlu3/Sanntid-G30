package master

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
)

// Arrays of all HallCalls and CabCalls
type Calls struct {
	HallCalls [config.N_FLOORS][config.N_BUTTONS - 1]bool
	CabCalls  [config.N_ELEVATORS][config.N_FLOORS]bool
}

// The messages sent between masters and backups
type BackupCalls struct {
	Calls Calls
	Id    int
}

// The message sent to the assigner
type AssignCalls struct {
	Calls          Calls
	AliveElevators [config.N_ELEVATORS]bool
}

// The added/removed calls we get from the slaveReceiver
type UpdateCalls struct {
	Calls   Calls
	AddCall bool
}

// struct{Calls Calls; Id int}
// struct{Calls Calls; AliveElevators [config.N_ELEVATORS]bool}
// struct{Calls Calls; AddCall bool}

/*
callsToBackupsTx transmitts calls read from callsToBackupsChan to the backups.
*/
func callsToBackupsTx(callsToBackupsChan <-chan Calls, initCalls Calls, Id int) {
	callsToBackupTransmitter := make(chan struct {
		Calls Calls
		Id    int
	})
	go bcast.Transmitter(config.MasterCallsPort, callsToBackupTransmitter)
	calls := initCalls
	for {
		select {
		case calls = <-callsToBackupsChan:
			callsToBackupTransmitter <- BackupCalls{Calls: calls, Id: Id}
		case <-time.After(config.MasterMessagePeriodSeconds):
			callsToBackupTransmitter <- BackupCalls{Calls: calls, Id: Id}
		}
	}

}

/*
fromAliveBackupsRxr listens the BackupsUpdatePort (from config) and if a backup is lost or reconnected, the function sends a list of the alive backups to aliveBackupsChan.
*/
func fromAliveBackupsRx(aliveBackupsChan chan<- []string) {
	updateFromBackupsChan := make(chan peers.PeerUpdate)
	go peers.Receiver(config.BackupsUpdatePort, updateFromBackupsChan)
	var aliveBackups []string
	for {
		update := <-updateFromBackupsChan
		fmt.Printf("Backups update:\n")
		fmt.Printf("  Backups:    %q\n", update.Peers)
		fmt.Printf("  New:        %q\n", update.New)
		fmt.Printf("  Lost:       %q\n", update.Lost)
		aliveBackups = update.Peers
		if len(update.Lost) != 0 || update.New != "" {
			aliveBackupsChan <- aliveBackups
		}
	}
}

/*
backupCoordinator starts goroutines to manage backup synchronization. It looks for other masters, tracks the status of alive backups, and sends call assignments to backups as needed.

This routine handles acknowledgments from alive backups, ensuring that all backups are synchronized with the current set of calls before turning the button lights on.
It also manages the reassignment of calls when necessary.
*/
func backupCoordinator(
	callsUpdateChan <-chan struct {
		Calls   Calls
		AddCall bool
	},
	callsToAssignChan chan<- struct {
		Calls          Calls
		AliveElevators [config.N_ELEVATORS]bool
	},
	calls Calls,
	Id int,
) {
	otherMasterUpdateChan := make(chan struct {
		Calls Calls
		Id    int
	})
	aliveBackupsChan := make(chan []string)
	callsToBackupsTxChan := make(chan Calls)
	callsFromBackupsRxChan := make(chan struct {
		Calls Calls
		Id    int
	})

	go bcast.Receiver(config.BackupsCallsPort, callsFromBackupsRxChan)
	go detectOtherMasters(otherMasterUpdateChan, Id)
	go fromAliveBackupsRx(aliveBackupsChan)
	go callsToBackupsTx(callsToBackupsTxChan, calls, Id)

	var aliveBackups []string
	var acksReceived [config.N_ELEVATORS]bool
	wantReassignment := true

mainLoop:
	for {
		select {
		case callsUpdate := <-callsUpdateChan:
			if callsUpdate.AddCall {
				calls = union(calls, callsUpdate.Calls)
			} else {
				calls = removeCalls(calls, callsUpdate.Calls)
			}
			callsToBackupsTxChan <- calls
			wantReassignment = true
			for i := range acksReceived {
				acksReceived[i] = false
			}
			acksReceived[Id] = true
		default:
		}

		select {
		case callsFromBackup := <-callsFromBackupsRxChan: // set ack for backup if it has the same calls
			if callsFromBackup.Calls == calls && !acksReceived[callsFromBackup.Id] {
				fmt.Println("new backup state from", callsFromBackup.Id)
				acksReceived[callsFromBackup.Id] = true
			}
		default:
		}

		select {
		case aliveBackups = <-aliveBackupsChan:
			wantReassignment = true
		default:
		}

		select {
		case otherMasterUpdate := <- otherMasterUpdateChan:
			if otherMasterUpdate.Id < Id && isCallsSubset(calls, otherMasterUpdate.Calls) {
				fmt.Println("find a better way to restart the program")
				os.Exit(42) // intentionally crashing, program restarts automatically when exiting with code 42
			} else if otherMasterUpdate.Id > Id {
				calls = union(calls, otherMasterUpdate.Calls)
				callsToBackupsTxChan <- calls
				wantReassignment = true
			} else {
				fmt.Println("couldn't end master phase: other master has not accepted our calls")
			}
		default:
		}

		for _, backup := range aliveBackups { // if some alive backups havent given ack, continue main loop
			backupId, _ := strconv.Atoi(backup)
			if !acksReceived[backupId] {
				continue mainLoop
			}
		}
		if wantReassignment {
			fmt.Println("BC: Sending calls")
			var aliveElevators [config.N_ELEVATORS]bool
			for _, backup := range aliveBackups {
				backupId, err := strconv.Atoi(backup)
				if err != nil {
					panic("BC got weird aliveElev")
				}
				aliveElevators[backupId] = true // if the backup is alive, then the elevator with the same id is alive
			}
			aliveElevators[Id] = true
			callsToAssignChan <- AssignCalls{Calls: calls, AliveElevators: aliveElevators}
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

// returns the set difference between calls and removedCalls
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
