package master

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type EventType int

const (
	Button       EventType = iota //In case of a button press
	FloorArrival                  //In case of a floor arrival
	Stuck                         //In case of a stuck elevator
)

// Translates to event for later use
type EventMessage struct {
	Elevator slave.Elevator     //The slave elevator
	Event    EventType          //The type of event
	Btn      elevio.ButtonEvent //Sends a button event
	Check    bool               //Sends a boolean for either stuck or not stuck
}

func receiveMessagesFromSlaves(slaveUpdate chan<- EventMessage) {
	for slaveID := 1; slaveID <= config.N_ELEVATORS; slaveID++ {
		go receiveMessageFromSlave(slaveUpdate, slaveID)
	}
}

func receiveMessageFromSlave(slaveUpdate chan<- EventMessage, slaveID int) {
	//rx channel for receiving from each slave
	rx := make(chan slave.EventMessage)
	go bcast.Receiver(config.SlaveBasePort+slaveID, rx)
	//ack channel to send an acknowledgment to each slave
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+slaveID+10, ack)
	for {
		select {}
	}
}

func sendMessagesToSlaves(slaveUpdate chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool) {
	tx := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	go bcast.Transmitter(config.SlaveBasePort, tx)

	var msg [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	for {
		select {
		case msg = <-slaveUpdate:
			tx <- msg
		default:
			tx <- msg
		}
	}
}

/*TODO:
-Fix what happens if a slave gets an order it immediately completes
-Consider an event for a slave clearing an order it sent
	-Solution to above ^ master should know the slave will clear said order*/
