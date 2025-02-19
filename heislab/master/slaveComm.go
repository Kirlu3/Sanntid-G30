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

func receiveMessagesFromSlaves(slaveUpdate chan EventMessage) {

	//Sends the rx channel to a receiver for each slave
	rx := make(chan slave.SlaveMessage)
	for i := 1; i <= config.N_ELEVATORS; i++ {
		go bcast.Receiver(config.SlaveBasePort+i, rx)
	}
	var prevElevator slave.Elevator
	var prevBtn elevio.ButtonEvent
	for {
		select {
		case msg := <-rx:
			if msg.PrevBtn != prevBtn {
				slaveUpdate <- EventMessage{Elevator: msg.Elevator, Event: Button, Btn: msg.PrevBtn, Check: false}
				prevBtn = msg.PrevBtn
			} else if msg.Elevator.Floor != prevElevator.Floor {
				slaveUpdate <- EventMessage{Elevator: msg.Elevator, Event: FloorArrival, Btn: elevio.ButtonEvent{}, Check: false}
			} else if msg.Elevator.Stuck != prevElevator.Stuck {
				slaveUpdate <- EventMessage{Elevator: msg.Elevator, Event: Stuck, Btn: elevio.ButtonEvent{}, Check: msg.Elevator.Stuck}
			}
		}
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
