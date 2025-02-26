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
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func Backup(id string) {

	// one time setup of communication between master and backup
	masterUpdateCh := make(chan peers.PeerUpdate)
	backupsUpdateCh := make(chan peers.PeerUpdate)
	masterTxEnable := make(chan bool)
	backupsTxEnable := make(chan bool)
	masterCallsTx := make(chan Slave.BackupCalls)
	backupCallsTx := make(chan Slave.BackupCalls)
	masterCallsRx := make(chan Slave.BackupCalls)
	backupCallsRx := make(chan Slave.BackupCalls)

	go peers.Transmitter(config.MasterUpdatePort, id, masterTxEnable)
	masterTxEnable <- false // this is dangerous as we risk briefly claiming to be master even though we are not, it seems as long as it takes less than interval it is fine
	go peers.Receiver(config.MasterUpdatePort, masterUpdateCh)

	go peers.Transmitter(config.BackupsUpdatePort, id, backupsTxEnable)
	go peers.Receiver(config.BackupsUpdatePort, backupsUpdateCh)

	go bcast.Transmitter(config.MasterWorldviewPort, masterCallsTx)
	go bcast.Transmitter(config.BackupsWorldviewPort, backupCallsTx)

	go bcast.Receiver(config.MasterWorldviewPort, masterCallsRx)
	go bcast.Receiver(config.BackupsWorldviewPort, backupCallsRx)

	fmt.Println("Backup Started: ", id)
	var backupsUpdate peers.PeerUpdate
	var masterUpdate peers.PeerUpdate
	var calls Slave.BackupCalls
	idInt, err := strconv.Atoi(id)
	if err != nil {
		panic("backup received invalid id")
	}
	calls.Id = idInt

	masterUpgradeCooldown := time.NewTimer(1 * time.Second)
	for {
		select {
		case c := <-masterCallsRx:
			if len(masterUpdate.Peers) > 0 && strconv.Itoa(c.Id) == masterUpdate.Peers[0] {
				calls.Calls = c.Calls
			} else {
				fmt.Println("received a message from not the master")
			}

		case backupsUpdate = <-backupsUpdateCh:
			fmt.Printf("Backups update:\n")
			fmt.Printf("  Backups:    %q\n", backupsUpdate.Peers)
			fmt.Printf("  New:        %q\n", backupsUpdate.New)
			fmt.Printf("  Lost:       %q\n", backupsUpdate.Lost)

		case masterUpdate = <-masterUpdateCh:
			fmt.Printf("Master update:\n")
			fmt.Printf("  Masters:    %q\n", masterUpdate.Peers)
			fmt.Printf("  New:        %q\n", masterUpdate.New)
			fmt.Printf("  Lost:       %q\n", masterUpdate.Lost)

		case <-time.After(time.Second * 1):
			fmt.Println("backup select blocked for 1 seconds. this should only happen if there are no masters")
		}
		backupCallsTx <- calls
		if len(masterUpdate.Peers) == 0 && len(backupsUpdate.Peers) != 0 && slices.Min(backupsUpdate.Peers) == id && func() bool {
			select {
			case <-masterUpgradeCooldown.C:
				return true
			default:
				return false
			}
		}() {
			backupsTxEnable <- false
			master.Master(calls, masterCallsTx, masterCallsRx, backupCallsRx, masterTxEnable, masterUpdateCh, backupsUpdateCh)
			backupsTxEnable <- true
			masterUpgradeCooldown.Reset(time.Second * 2)
		}
	}
}
