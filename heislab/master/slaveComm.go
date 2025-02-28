package master

import (
	"fmt"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func receiveMessagesFromSlaves(stateUpdateCh chan<- slave.Elevator, callsUpdateCh chan<- slave.UpdateCalls) {
	for slaveID := range config.N_ELEVATORS {
		go receiveMessageFromSlave(stateUpdateCh, callsUpdateCh, slaveID)
	}
}

func receiveMessageFromSlave(stateUpdateCh chan<- slave.Elevator, callsUpdateCh chan<- slave.UpdateCalls, slaveID int) {
	//rx channel for receiving from each slave
	rx := make(chan slave.EventMessage)
	go bcast.Receiver(config.SlaveBasePort+slaveID, rx)
	//ack channel to send an acknowledgment to each slave
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+slaveID+10, ack)
	var msgID int
	for msg := range rx {
		println("ST: Received message")
		ack <- msg.MsgID
		fmt.Println("ST: Sent Ack")
		if msg.MsgID != msgID {
			println("ST: Received new message")
			msgID = msg.MsgID
			stateUpdateCh <- msg.Elevator
			// put all of this into a function maybe??
			if msg.Event == slave.Button {
				callsUpdate := makeAddCallsUpdate(msg)
				callsUpdateCh <- callsUpdate
			} else if msg.Event == slave.FloorArrival && msg.Elevator.Behaviour == slave.EB_DoorOpen {
				callsUpdate := makeRemoveCallsUpdate(msg)
				callsUpdateCh <- callsUpdate
			}
			println("ST: Sent message out")
		}
	}
}

// TODO fix logic for removing hall calls, because it doesnt really make any sense to me
func makeRemoveCallsUpdate(msg slave.EventMessage) slave.UpdateCalls {
	var callsUpdate slave.UpdateCalls
	callsUpdate.AddCall = false
	callsUpdate.Calls.CabCalls[msg.Elevator.ID][msg.Elevator.Floor] = true
	callsUpdate.Calls.HallCalls[msg.Elevator.Floor][0] = true
	callsUpdate.Calls.HallCalls[msg.Elevator.Floor][1] = true
	if msg.Elevator.Floor == 0 {
		callsUpdate.Calls.HallCalls[0][elevio.BT_HallDown] = false
	} else if msg.Elevator.Floor == config.N_FLOORS-1 {
		callsUpdate.Calls.HallCalls[config.N_FLOORS-1][elevio.BT_HallUp] = false
	} else if msg.Elevator.Direction == slave.D_Down {
		callsUpdate.Calls.HallCalls[msg.Elevator.Floor][elevio.BT_HallUp] = false
	} else if msg.Elevator.Direction == slave.D_Up {
		callsUpdate.Calls.HallCalls[msg.Elevator.Floor][elevio.BT_HallDown] = false
	}
	return callsUpdate
}

func makeAddCallsUpdate(msg slave.EventMessage) slave.UpdateCalls {
	var callsUpdate slave.UpdateCalls
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
