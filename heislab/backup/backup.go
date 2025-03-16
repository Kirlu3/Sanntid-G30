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
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

/*
	The entire backup system run in one goroutine.

The routine listens to the master's UDP broadcasts and responds with the updated worldview.
If the backup loses connection with the master, it will transition to the master phase with its current worldview.
A large portion of the backup code are pretty prints of updates to peer lists.
*/
func Backup(
	id string,
	masterToSlaveCalls_offlineChan chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	slaveToMasterBtn_offlineChan <-chan slave.ButtonMessage,
	slaveToMasterElevState_offlineChan <-chan slave.Elevator,
) {
	masterUpdateChan := make(chan peers.PeerUpdate)
	backupsUpdateChan := make(chan peers.PeerUpdate)
	backupsTransmitterEnableChan := make(chan bool)
	backupsCallsTransmitterChan := make(chan struct {
		Calls master.Calls
		Id    int
	})
	mastersCallsReceiverChan := make(chan struct {
		Calls master.Calls
		Id    int
	})

	go peers.Receiver(config.MasterUpdatePort, masterUpdateChan)

	go peers.Transmitter(config.BackupsUpdatePort, id, backupsTransmitterEnableChan)
	go peers.Receiver(config.BackupsUpdatePort, backupsUpdateChan)

	go bcast.Transmitter(config.BackupsCallsPort, backupsCallsTransmitterChan)

	go bcast.Receiver(config.MasterCallsPort, mastersCallsReceiverChan)

	fmt.Println("Backup Started: ", id)
	var backupsUpdate peers.PeerUpdate
	var masterUpdate peers.PeerUpdate
	var calls master.Calls

	idInt, err := strconv.Atoi(id)
	if err != nil {
		panic("backup received invalid id")
	}

	masterUpdateCooldownTimer := time.NewTimer(1 * time.Second)

	for {
		select {
		case c := <-mastersCallsReceiverChan:
			if len(masterUpdate.Peers) > 0 && strconv.Itoa(c.Id) == masterUpdate.Peers[0] {
				calls = c.Calls
			} else {
				fmt.Println("received a message from not the master")
			}

		case backupsUpdate = <-backupsUpdateChan:
			fmt.Printf("Backups update:\n")
			fmt.Printf("  Backups:    %q\n", backupsUpdate.Peers)
			fmt.Printf("  New:        %q\n", backupsUpdate.New)
			fmt.Printf("  Lost:       %q\n", backupsUpdate.Lost)

		case masterUpdate = <- masterUpdateChan:
			fmt.Printf("Master update:\n")
			fmt.Printf("  Masters:    %q\n", masterUpdate.Peers)
			fmt.Printf("  New:        %q\n", masterUpdate.New)
			fmt.Printf("  Lost:       %q\n", masterUpdate.Lost)

		case <-time.After(time.Second * 2):
			fmt.Println("backup select blocked for 2 seconds. this should only happen if there are no masters, maybe this is too short?")
		}
		backupsCallsTransmitterChan <- master.BackupCalls{Calls: calls, Id: idInt}
		if len(masterUpdate.Peers) == 0 && len(backupsUpdate.Peers) != 0 && slices.Min(backupsUpdate.Peers) == id && func() bool {
			select {
			case <-masterUpdateCooldownTimer.C:
				return true
			default:
				return false
			}
		}() {
			backupsTransmitterEnableChan <- false
			master.Master(calls, idInt,  masterToSlaveCalls_offlineChan, slaveToMasterBtn_offlineChan, slaveToMasterElevState_offlineChan)
			panic("the master phase should never return")
		}
	}
}
