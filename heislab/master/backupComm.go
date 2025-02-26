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

func backupsTx(stateToBackup <-chan slave.WorldView, masterCallsTx chan<- slave.BackupCalls, initCalls slave.BackupCalls) {
	calls := deepcopy.Copy(initCalls).(slave.BackupCalls)
	for {
		select {
		case worldview := <-stateToBackup:
			calls.Calls.CabCalls = worldview.CabCalls
			calls.Calls.HallCalls = worldview.HallCalls
			masterCallsTx <- calls
		case <-time.After(config.MasterMessagePeriodSeconds):
			masterCallsTx <- calls
		}
	}

}

func aliveBackupsRx(aliveBackupsCh chan<- []string, backupsUpdateCh <-chan peers.PeerUpdate) {
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
func receiveBackupAck(OwnId string, requestBackupAckCh <-chan slave.Calls, aliveBackupsCh <-chan []string, aliveBackupsToManagerCh chan<- []string, callsToAssign chan<- slave.AssignCalls, backupCallsRx <-chan slave.BackupCalls) {
	ID, _ := strconv.Atoi(OwnId)
	var aliveBackups []string = <-aliveBackupsCh
	aliveBackupsToManagerCh <- aliveBackups
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
		case a := <-backupCallsRx: // set ack for backup if it has the same calls
			if a.Calls == calls {
				fmt.Println("new backup state from", a.Id)
				acksReceived[a.Id] = true
			}
		default:
		}

		select {
		case aliveBackups = <-aliveBackupsCh:
			// aliveBackupsToManagerCh <- aliveBackups // oh wtf BEGGING FOR A DEADLOCK // SOLUTION: DONT SEND ALIVE TO STATEMANAGER
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
			var AliveElevators [config.N_ELEVATORS]bool
			for _, elev := range aliveBackups {
				idx, err := strconv.Atoi(elev)
				if err != nil {
					panic("BC got weird aliveElev")
				}
				AliveElevators[idx] = true
			}
			AliveElevators[ID] = true
			callsToAssign <- slave.AssignCalls{Calls: calls, AliveElevators: AliveElevators} // orders to assign
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
