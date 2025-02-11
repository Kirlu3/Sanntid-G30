package Backup

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

// maybe make a config file where we can change all parameters
// set the time until it is time to become the master
// i guess each node should have an id and associated address/PORT e.g. id = x, port = 6x009?

// find a way to init the program with an id

// we should be operating with the

// network module decides this
// var _BACKUP_TAKEOVER_SECONDS int = 5

// information the backup needs that is not stored in the Elevator object
type ElevatorMeta struct {
	id int
	isMaster int
}

type ExpandedElevator struct {
	elevator Slave.Elevator
	meta ElevatorMeta
}

	
// should this actually be a struct?
type WorldView struct {
	expandedElevators []ExpandedElevator
	myId int
}


func Backup(id int) {
	var worldView WorldView = init_unknown_world_view()
	var peerUpdate peers.PeerUpdate


	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	worldViewTx := make(chan WorldView)
	worldViewRx := make(chan WorldView)

	go bcast.Transmitter(16569, worldViewTx)
	go bcast.Receiver(16569, worldViewRx)

	for {
		// what is our local state on startup?
		// init to DONT_KNOW


		// send my worldview periodically
		go func() {
			for {
				worldViewTx <- worldView
				time.Sleep(1 * time.Second) // how often is message sent?
			}
		}()


		fmt.Println("Started")
		messageHandlerLoop:
		for {
			select {
			case peerUpdate = <-peerUpdateCh:
				fmt.Printf("Peer update:\n")
				fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
				fmt.Printf("  New:      %q\n", peerUpdate.New)
				fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
				// break if no master is sending messages
				// break if we are the only one on the network? because clearly then we should become master?
				// maybe we just loop through everyone and 
				if (peer) {
					break messageHandlerLoop
				}
	
			case a := <-worldViewRx:
				fmt.Printf("Received: %#v\n", a)
				if (is_master_elevator(a)) {
					worldView.expandedElevators = a.expandedElevators
				}
			}
		}

		// close the old channels?
		if (lowest_id(peerUpdate.Peers, id)) {
			Master.Master(full state i guess)	
		}
	}
}

