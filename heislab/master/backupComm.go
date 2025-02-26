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

func backupsTx(callsToBackupsCh <-chan slave.Calls, masterCallsTx chan<- slave.BackupCalls, initCalls slave.BackupCalls) {
	calls := deepcopy.Copy(initCalls).(slave.BackupCalls)
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
	callsUpdateCh <-chan slave.Calls, //the message we get from slaveRx to calculate updated calls are doesnt have to be this type, but with this + calls we should be able to calculate what the updated calls should be
	callsToAssignCh chan<- slave.AssignCalls,
	endMasterPhaseCh chan<- struct{},
	initCalls slave.BackupCalls,
	masterCallsTx chan<- slave.BackupCalls,
	masterCallsRx <-chan slave.BackupCalls,
	backupCallsRx <-chan slave.BackupCalls,
	backupsUpdateCh <-chan peers.PeerUpdate,
) {
	Id := initCalls.Id

	otherMasterCallsCh := make(chan slave.BackupCalls)
	aliveBackupsCh := make(chan []string)
	callsToBackupsCh := make(chan slave.Calls)

	// for all channels consider if it is nicer to start them here or in master()
	go lookForOtherMasters(otherMasterCallsCh, Id, masterCallsRx)
	go aliveBackupsRx(aliveBackupsCh, backupsUpdateCh)
	go backupsTx(callsToBackupsCh, masterCallsTx, deepcopy.Copy(initCalls).(slave.BackupCalls))

	var aliveBackups []string
	var acksReceived [config.N_ELEVATORS]bool
	calls := deepcopy.Copy(initCalls).(slave.Calls)
	wantReassignment := false
mainLoop:
	for {
		select {
		case callsUpdate := <-callsUpdateCh: // when we receive new calls reset all acks, THE LOGIC HERE WILL BE MORE COMPLICATED NOW
			calls = updatedCalls(calls, callsUpdate) //TODO: calculate the updated calls based on calls and the new message
			wantReassignment = true
			for i := range acksReceived {
				acksReceived[i] = false
			}
			acksReceived[Id] = true
		default:
		}

		select {
		case a := <-backupCallsRx: // set ack for backup if it has the same calls
			if a.Calls == calls {
				// fmt.Println("new backup state from", a.Id)
				acksReceived[a.Id] = true
			}
		default:
		}

		select {
		case aliveBackups = <-aliveBackupsCh:
			wantReassignment = true
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
			callsToAssignCh <- slave.AssignCalls{Calls: calls, AliveElevators: AliveElevators} // orders to assign, do we ever block here?
			wantReassignment = false
		}
	}
}
