package backup

import (
	"context"
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

	for {
		backupsTxEnable <- true
		fmt.Println("Backup Started: ", id)
		var backupsUpdate peers.PeerUpdate
		var masterUpdate peers.PeerUpdate
		ctx, cancel := context.WithCancel(context.Background())
		callsUpdateCh := make(chan Slave.Calls)

		go func() {
			var calls Slave.BackupCalls
			idInt, err := strconv.Atoi(id)
			if err != nil {
				panic("backup received invalid id")
			}
			calls.Id = idInt

			for {
				var ok bool
				select {
				case calls.Calls, ok = <-callsUpdateCh:
					if !ok {
						// make sure our calls are brought into the master phase
						return
					}

				default:
					backupCallsTx <- calls
				}
			}
		}()

		go func() {
			var calls Slave.BackupCalls
			idInt, err := strconv.Atoi(id)
			if err != nil {
				panic("backup received invalid id")
			}
			calls.Id = idInt
			for {
				select {
				case <-ctx.Done():
					close(callsUpdateCh)
					return

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
				}
				backupCallsTx <- calls
			}
		}()

		time.Sleep(time.Millisecond*200)
		for {
			if slices.Min(backupsUpdate.Peers) == id {
				cancel()
				// close the old channels? it might not be strictly necessary, // TODO fix
				backupsTxEnable <- false // consider this
				master.Master(worldView, masterUpdateCh, masterTxEnable, masterWorldViewTx, masterWorldViewRx, backupWorldViewRx, backupsUpdateCh)
			}
		}


		// start master phase:
		

	// messageHandlerLoop:
	// 	for {
	// 		select {
	// 		case <-time.After(2 * time.Second):
	// 			//does this still run if you activate another case?
	// 			break messageHandlerLoop
	// 		case masterUpdate = <-masterUpdateCh:
	// 			fmt.Printf("Master update:\n")
	// 			fmt.Printf("  Masters:    %q\n", masterUpdate.Peers)
	// 			fmt.Printf("  New:        %q\n", masterUpdate.New)
	// 			fmt.Printf("  Lost:       %q\n", masterUpdate.Lost)
	// 			if len(masterUpdate.Peers) == 0 {
	// 				break messageHandlerLoop
	// 			}

	// 		case backupsUpdate = <-backupsUpdateCh:
	// 			fmt.Printf("Backups update:\n")
	// 			fmt.Printf("  Backups:    %q\n", backupsUpdate.Peers)
	// 			fmt.Printf("  New:        %q\n", backupsUpdate.New)
	// 			fmt.Printf("  Lost:       %q\n", backupsUpdate.Lost)

	// 		case a := <-masterWorldViewRx:
	// 			// fmt.Printf("Received: %#v\n", a)
	// 			if len(masterUpdate.Peers) > 0 && a.OwnId == masterUpdate.Peers[0] {
	// 				worldView.CabCalls = a.CabCalls
	// 				worldView.HallCalls = a.HallCalls

	// 			} else {
	// 				fmt.Println("received master state from not the master")
	// 			}
	// 		}
	// 	}


	}
}
