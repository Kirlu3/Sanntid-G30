package master

import (
	"os"
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/alive"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
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
# Broadcasts all active calls to the backups

Receives calls on the callsToBackupsChan and broadcasts on the MasterBroadcastPort
*/
func callsToBackupsTx(callsToBackupsChan <-chan Calls, initCalls Calls, Id int) {
	callsToBackupTx := make(chan struct {
		Calls Calls
		Id    int
	})
	go bcast.Transmitter(config.MasterBroadcastPort, callsToBackupTx)
	calls := initCalls
	for {
		select {
		case calls = <-callsToBackupsChan:
			callsToBackupTx <- BackupCalls{Calls: calls, Id: Id}
		case <-time.After(config.MasterBroadcastCallsPeriodMs * time.Millisecond):
			callsToBackupTx <- BackupCalls{Calls: calls, Id: Id}
		}
	}
}

/*
# Listens to both the backup and master broadcast ports and ensures acknowledgments on all active calls

Will crash the program upon encountering another master of higher priority
*/
func callsFromBackupsRx(
	callsUpdateChan <-chan struct {
		Calls   Calls
		AddCall bool
	},
	callsToAssignChan chan<- struct {
		Calls          Calls
		AliveElevators [config.N_ELEVATORS]bool
	},
	callsToBackupsTxChan chan<- Calls,
	initCalls Calls,
	ownId int,
) {
	masterBroadcastRxChan := make(chan struct {
		Calls Calls
		Id    int
	})
	backupBroadcastRxChan := make(chan struct {
		Calls Calls
		Id    int
	})
	aliveBackupsUpdateChan := make(chan alive.AliveUpdate)

	go bcast.Receiver(config.MasterBroadcastPort, masterBroadcastRxChan)
	go bcast.Receiver(config.BackupsBroadcastPort, backupBroadcastRxChan)
	go alive.Receiver(config.BackupsUpdatePort, aliveBackupsUpdateChan)

	var aliveBackups []string
	var acksReceived [config.N_ELEVATORS]bool

	calls := initCalls
	callsToBackupsTxChan <- calls

mainLoop:
	for {
		select {
		case callsUpdate := <-callsUpdateChan:
			calls, acksReceived = incomingCallsUpdate(calls, acksReceived, ownId, callsUpdate)
			callsToBackupsTxChan <- calls

		case backupBroadcast := <-backupBroadcastRxChan: // set ack for backup if it has the same calls
			acksReceived = incomingBackupBroadcast(calls, acksReceived, backupBroadcast)

		case aliveBackupsUpdate := <-aliveBackupsUpdateChan:
			aliveBackups = aliveBackupsUpdate.Alive

		case masterBroadcast := <-masterBroadcastRxChan:
			calls = incomingMasterBroadcast(calls, ownId, masterBroadcast)
			callsToBackupsTxChan <- calls
		case <-time.After(config.MasterBroadcastCallsPeriodMs * time.Millisecond):
		}

		for _, backup := range aliveBackups { // if some alive backups havent given ack, continue main loop
			backupId, _ := strconv.Atoi(backup)
			if !acksReceived[backupId] {
				continue mainLoop
			}
		}

		var aliveElevators [config.N_ELEVATORS]bool
		for _, backup := range aliveBackups {
			backupId, _ := strconv.Atoi(backup)
			aliveElevators[backupId] = true // if the backup is alive, then the elevator with the same id is alive
		}
		aliveElevators[ownId] = true
		callsToAssignChan <- AssignCalls{Calls: calls, AliveElevators: aliveElevators}
	}
}

/*
# Called when there is an incoming calls update

# Updates the calls and reset acksReceived if there was a change in calls

Input: the active calls, the active acksReceived, the master's ID, and the callsUpdate to be processed

Returns: updated calls and acksReceived
*/
func incomingCallsUpdate(calls Calls, acksReceived [config.N_ELEVATORS]bool, ownID int, callsUpdate struct {
	Calls   Calls
	AddCall bool
}) (Calls, [config.N_ELEVATORS]bool) {

	var newCalls Calls
	if callsUpdate.AddCall {
		newCalls = addCalls(calls, callsUpdate.Calls)
	} else {
		newCalls = removeCalls(calls, callsUpdate.Calls)
	}

	if calls != newCalls {
		calls = newCalls

		for i := range acksReceived {
			acksReceived[i] = false
		}
		acksReceived[ownID] = true
	}
	return calls, acksReceived
}

/*
# Called when there is an incoming broadcast on the BackupBroadcastPort

# updates acknowledgements if the calls match

Input: the active calls, the active acksReceived, the incoming backupBroadcast

Returns: updated acksReceived
*/
func incomingBackupBroadcast(calls Calls, acksReceived [config.N_ELEVATORS]bool, backupBroadcast struct {
	Calls Calls
	Id    int
}) [config.N_ELEVATORS]bool {
	if backupBroadcast.Calls == calls && !acksReceived[backupBroadcast.Id] {
		acksReceived[backupBroadcast.Id] = true
	}
	return acksReceived
}

/*
# Called when a broadcast is received on the master port

If the master hears itself: does nothing

If it hears a master with lower priority: add it's calls

If it hears a master with higher priority: die once it knows it's calls have been received

Input: the active calls, the master's ID, and the incoming masterBroadcast

Returns: updated calls
*/
func incomingMasterBroadcast(calls Calls, ownID int, masterBroadcast struct {
	Calls Calls
	Id    int
}) Calls {

	if masterBroadcast.Id == ownID {
		//Do nothing
	} else if masterBroadcast.Id > ownID {
		calls = addCalls(calls, masterBroadcast.Calls)

	} else if masterBroadcast.Id < ownID && isCallsSubset(calls, masterBroadcast.Calls) {
		os.Exit(42) // intentionally crashing, program restarts automatically when exiting with code 42
	}
	return calls
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
func addCalls(calls1 Calls, calls2 Calls) Calls {
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
