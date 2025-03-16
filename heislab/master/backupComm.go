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
masterToBackupsTransmitter transmitts calls read from callsToBackupsChan to the backups
*/
func masterToBackupsTransmitter(callsToBackupsChan <-chan Calls, initCalls Calls, Id int) {
	masterCallsTx := make(chan struct {
		Calls Calls
		Id    int
	})
	go bcast.Transmitter(config.MasterCallsPort, masterCallsTx)
	calls := initCalls
	for {
		select {
		case calls = <-callsToBackupsChan:
			masterCallsTx <- BackupCalls{Calls: calls, Id: Id}
		case <-time.After(config.MasterMessagePeriodSeconds):
			masterCallsTx <- BackupCalls{Calls: calls, Id: Id}
		}
	}

}

/*
fromAliveBackupsReceiver listens the BackupsUpdatePort (from config) and if a backup is lost or reconnected, the function sends a list of the aliveBackups to aliveBackupsChan.
*/
func fromAliveBackupsReceiver(aliveBackupsChan chan<- []string) {
	backupsUpdateChan := make(chan peers.PeerUpdate)
	go peers.Receiver(config.BackupsUpdatePort, backupsUpdateChan)
	var aliveBackups []string
	for {
		a := <-backupsUpdateChan
		fmt.Printf("Backups update:\n")
		fmt.Printf("  Backups:    %q\n", a.Peers)
		fmt.Printf("  New:        %q\n", a.New)
		fmt.Printf("  Lost:       %q\n", a.Lost)
		aliveBackups = a.Peers
		if len(a.Lost) != 0 || a.New != "" {
			aliveBackupsChan <- aliveBackups
		}
	}
}

/*
backupAckReceiver starts goroutines to manage backup synchronization. It looks for other masters, tracks the status of alive backups, and sends call assignments to backups as needed.

This routine handles acknowledgments from alive backups, ensuring that all backups are synchronized with the current set of calls before turning the button lights on.
It also manages the reassignment of calls when necessary.
*/
func backupAckReceiver(
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
	updateFromOtherMasterChan := make(chan struct {
		Calls Calls
		Id    int
	})
	aliveBackupsChan := make(chan []string)
	callsToBackupsChan := make(chan Calls)
	backupCallsReceiver := make(chan struct {
		Calls Calls
		Id    int
	})

	go bcast.Receiver(config.BackupsCallsPort, backupCallsReceiver)
	go lookForOtherMasters(updateFromOtherMasterChan, Id)
	go fromAliveBackupsReceiver(aliveBackupsChan)
	go masterToBackupsTransmitter(callsToBackupsChan, calls, Id)

	var aliveBackups []string
	var acksReceived [config.N_ELEVATORS]bool
	wantReassignment := true //why?
mainLoop:
	for {
		select {
		case callsUpdate := <-callsUpdateChan:
			if callsUpdate.AddCall {
				calls = union(calls, callsUpdate.Calls)
			} else {
				calls = removeCalls(calls, callsUpdate.Calls)
			}
			callsToBackupsChan <- calls
			wantReassignment = true
			for i := range acksReceived {
				acksReceived[i] = false
			}
			acksReceived[Id] = true
		default:
		}

		select {
		case a := <-backupCallsReceiver: // set ack for backup if it has the same calls
			if a.Calls == calls && !acksReceived[a.Id] {
				fmt.Println("new backup state from", a.Id)
				acksReceived[a.Id] = true
			}
		default:
		}

		select {
		case aliveBackups = <-aliveBackupsChan:
			wantReassignment = true
		default:
		}

		select {
		case updateFromOtherMaster := <-updateFromOtherMasterChan:
			if updateFromOtherMaster.Id < Id && isCallsSubset(calls, updateFromOtherMaster.Calls) {
				fmt.Println("find a better way to restart the program")
				os.Exit(42) // intentionally crashing, program restarts automatically when exiting with code 42
			} else if updateFromOtherMaster.Id > Id {
				calls = union(calls, updateFromOtherMaster.Calls)
				callsToBackupsChan <- calls
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
			var aliveElevators [config.N_ELEVATORS]bool
			for _, id := range aliveBackups {
				id_Int, err := strconv.Atoi(id)
				if err != nil {
					panic("BC got weird aliveElev")
				}
				aliveElevators[id_Int] = true
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
