package backup

import (
	"fmt"
	"slices"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func Backup(id string) {
	var worldView Slave.WorldView
	worldView.OwnId = id
	var backupsUpdate peers.PeerUpdate
	var masterUpdate peers.PeerUpdate

	backupsUpdateCh := make(chan peers.PeerUpdate)
	backupsTxEnable := make(chan bool)

	go peers.Transmitter(config.BackupsUpdatePort, id, backupsTxEnable)
	go peers.Receiver(config.BackupsUpdatePort, backupsUpdateCh)

	masterUpdateCh := make(chan peers.PeerUpdate)
	masterTxEnable := make(chan bool)

	go peers.Transmitter(config.MasterUpdatePort, id, masterTxEnable)
	masterTxEnable <- false // this is dangerous as we risk briefly claiming to be master even though we are not, it seems as long as it takes less than interval it is fine
	go peers.Receiver(config.MasterUpdatePort, masterUpdateCh)

	backupWorldViewTx := make(chan Slave.WorldView)
	masterWorldViewRx := make(chan Slave.WorldView)

	go bcast.Transmitter(config.BackupsWorldviewPort, backupWorldViewTx)
	go bcast.Receiver(config.MasterWorldviewPort, masterWorldViewRx)

	// init for the master phase
	backupWorldViewRx := make(chan Slave.WorldView)
	masterWorldViewTx := make(chan Slave.WorldView)

	go bcast.Receiver(config.BackupsWorldviewPort, backupWorldViewRx)
	go bcast.Transmitter(config.MasterWorldviewPort, masterWorldViewTx)

	for {
		backupsTxEnable <- true
		// send my worldview periodically, should we stop this when we become master? or just create one that runs forever
		go func() {
			for {
				backupWorldViewTx <- worldView
				time.Sleep(config.BackupMessagePeriodSeconds * time.Second) // how often is message sent?
			}
		}()

		fmt.Println("Started")
	messageHandlerLoop:
		for {
			select {
			case masterUpdate = <-masterUpdateCh:
				fmt.Printf("Master update:\n")
				fmt.Printf("  Masters:    %q\n", masterUpdate.Peers)
				fmt.Printf("  New:        %q\n", masterUpdate.New)
				fmt.Printf("  Lost:       %q\n", masterUpdate.Lost)
				if len(masterUpdate.Peers) == 0 {
					break messageHandlerLoop
				}

			// case backupsUpdate = <-backupsUpdateCh:
			// fmt.Printf("Backups update:\n")
			// fmt.Printf("  Backups:    %q\n", backupsUpdate.Peers)
			// fmt.Printf("  New:        %q\n", backupsUpdate.New)
			// fmt.Printf("  Lost:       %q\n", backupsUpdate.Lost)

			case a := <-masterWorldViewRx:
				fmt.Printf("Received: %#v\n", a)
				if len(masterUpdate.Peers) > 0 && a.OwnId == masterUpdate.Peers[0] {
					worldView.Elevators = a.Elevators // i have no idea if this is ok or if we get shallow copy problems with slices
				} else {
					fmt.Println("received master state from not the master")
				}
			}
		}

		if min(slices.Min(backupsUpdate.Peers)) == id {
			// close the old channels? it might not be strictly necessary, // TODO fix
			backupsTxEnable <- false // consider this
			master.Master(worldView, masterUpdateCh, masterTxEnable, masterWorldViewTx, masterWorldViewRx, backupWorldViewRx, backupsUpdateCh)
		}
	}
}
