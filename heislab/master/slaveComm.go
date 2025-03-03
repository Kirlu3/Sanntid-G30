package master

import (
	"fmt"
	"slices"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

// how do I clear orders?
func receiveMessagesFromSlaves(stateUpdateCh chan<- slave.Elevator,
	callsUpdateCh chan<- UpdateCalls,
	assignmentsToSlaveReceiver <-chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool) {

	slaveRx := make(chan slave.EventMessage)
	for slaveID := range config.N_ELEVATORS {
		go receiveMessageFromSlave(slaveRx, slaveID)
	}

	var assignments [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	for {
		select {
		case update := <-slaveRx:
			fmt.Println("ST: Received new message")
			fmt.Println(update)
			switch update.Event {
			case slave.Button:
				callsUpdateCh <- makeAddCallsUpdate(update)
			case slave.FloorArrival:
				stateUpdateCh <- update.Elevator
				if update.Elevator.Behaviour == slave.EB_DoorOpen {
					callsUpdateCh <- makeRemoveCallsUpdate(update, assignments)
				}
			case slave.Stuck:
				stateUpdateCh <- update.Elevator
			}
		case assignments = <-assignmentsToSlaveReceiver:
			continue
		}
	}
}

func receiveMessageFromSlave(slaveRx chan<- slave.EventMessage, slaveID int) {

	//rx channel for receiving from each slave
	rx := make(chan slave.EventMessage)
	go bcast.Receiver(config.SlaveBasePort+slaveID, rx)
	//ack channel to send an acknowledgment to each slave
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+slaveID+10, ack)

	var msgID []int
	for msg := range rx {
		println("ST: Received message")
		ack <- msg.MsgID
		fmt.Println("ST: Sent Ack", msg.MsgID)
		if !slices.Contains(msgID, msg.MsgID) {
			msgID = append(msgID, msg.MsgID)
			// if we've stored too many IDs, remove the oldest one (I assume I will never need to hold more than 10, likely less)
			if len(msgID) > 10 {
				msgID = msgID[1:]
			}
			slaveRx <- msg
		}
	}
}

// TODO fix logic for removing hall calls, because it doesnt really make any sense to me
func makeRemoveCallsUpdate(msg slave.EventMessage, assignments [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool) UpdateCalls {
	var callsUpdate UpdateCalls
	callsUpdate.AddCall = false

	callsUpdate.Calls.CabCalls[msg.Elevator.ID][msg.Elevator.Floor] = true
	for btn := range config.N_BUTTONS - 1 {
		if assignments[msg.Elevator.ID][msg.Elevator.Floor][btn] && !msg.Elevator.Requests[msg.Elevator.Floor][btn] {
			callsUpdate.Calls.HallCalls[msg.Elevator.Floor][btn] = true
		}
	}
	return callsUpdate
}

func makeAddCallsUpdate(msg slave.EventMessage) UpdateCalls {
	var callsUpdate UpdateCalls
	callsUpdate.AddCall = true
	if msg.Btn.Button == elevio.BT_Cab {
		callsUpdate.Calls.CabCalls[msg.Elevator.ID][msg.Btn.Floor] = true
	} else {
		callsUpdate.Calls.HallCalls[msg.Btn.Floor][msg.Btn.Button] = true
	}
	return callsUpdate
}

func sendMessagesToSlaves(toSlaveCh chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool) {
	tx := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	go bcast.Transmitter(config.SlaveBasePort-1, tx)

	var msg [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	for {
		//Gives message frequency
		time.Sleep(time.Millisecond * 5)

		select {
		case msg = <-toSlaveCh:
			fmt.Println("ST: New orders sent")
			fmt.Println(msg)
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
