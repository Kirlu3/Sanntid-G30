package backup

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
)

/*
	The entire backup system run in one goroutine.

The routine listens to the master's UDP broadcasts and responds with the updated calls.
If the backup loses connection with the master, it will transition to the master phase with its current list of calls.
A large portion of the backup code are pretty prints of updates to peer lists.
*/
func Backup(id string) master.Calls {
	masterUpdateRxChan := make(chan peers.PeerUpdate)
	backupsUpdateRxChan := make(chan peers.PeerUpdate)
	enableBackupTxChan := make(chan bool)
	backupCallsTxChan := make(chan struct {
		Calls master.Calls
		Id    int
	})
	masterCallsRxChan := make(chan struct {
		Calls master.Calls
		Id    int
	})

	go peers.Receiver(config.MasterUpdatePort, masterUpdateRxChan)

	go peers.Transmitter(config.BackupsUpdatePort, id, enableBackupTxChan)
	go peers.Receiver(config.BackupsUpdatePort, backupsUpdateRxChan)

	go bcast.Transmitter(config.BackupsCallsPort, backupCallsTxChan)

	go bcast.Receiver(config.MasterCallsPort, masterCallsRxChan)

	fmt.Println("Backup Started: ", id)
	var backupsUpdate peers.PeerUpdate
	var masterUpdate peers.PeerUpdate
	var calls master.Calls

	idInt, err := strconv.Atoi(id)
	if err != nil {
		panic("backup received invalid id")
	}

	masterUpgradeCooldownTimer := time.NewTimer(1 * time.Second)

	for {
		select {
		case newCalls := <-masterCallsRxChan:
			if len(masterUpdate.Peers) > 0 && strconv.Itoa(newCalls.Id) == masterUpdate.Peers[0] {
				calls = newCalls.Calls
			} else {
				fmt.Println("received a message from not the master")
			}

		case backupsUpdate = <-backupsUpdateRxChan:
			fmt.Printf("Backups update:\n")
			fmt.Printf("  Backups:    %q\n", backupsUpdate.Peers)
			fmt.Printf("  New:        %q\n", backupsUpdate.New)
			fmt.Printf("  Lost:       %q\n", backupsUpdate.Lost)

		case masterUpdate = <-masterUpdateRxChan:
			fmt.Printf("Master update:\n")
			fmt.Printf("  Masters:    %q\n", masterUpdate.Peers)
			fmt.Printf("  New:        %q\n", masterUpdate.New)
			fmt.Printf("  Lost:       %q\n", masterUpdate.Lost)

		case <-time.After(time.Second * 2):
			fmt.Println("No new messages for two seconds, no master available")
		}
		backupCallsTxChan <- master.BackupCalls{Calls: calls, Id: idInt}
		if len(masterUpdate.Peers) == 0 && (len(backupsUpdate.Peers) == 0 || slices.Min(backupsUpdate.Peers) == id) {
			select {
			case <-masterUpgradeCooldownTimer.C:
				enableBackupTxChan <- false
				return calls
			default:
			}
		}

	}
}
