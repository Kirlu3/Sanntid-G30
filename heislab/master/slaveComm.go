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
func receiveMessagesFromSlaves(stateUpdateCh chan<- [config.N_ELEVATORS]slave.Elevator, callsUpdateCh chan<- slave.Calls) {
	slaveRx := make(chan slave.EventMessage)
	for slaveID := range config.N_ELEVATORS {
		go receiveMessageFromSlave(slaveRx, slaveID)
	}

	for update := range slaveRx {
		fmt.Println("ST: Received new message")
		fmt.Println(update)
		switch update.Event {
		case slave.Button:
			var newCalls slave.Calls
			if update.Btn.Button == elevio.BT_Cab {
				newCalls.CabCalls[update.Elevator.ID][update.Btn.Floor] = true
			} else {
				newCalls.HallCalls[update.Btn.Floor][update.Btn.Button] = true
			}
			callsUpdateCh <- newCalls
		case slave.FloorArrival:
			var newStates [config.N_ELEVATORS]slave.Elevator
			newStates[update.Elevator.ID] = update.Elevator
			stateUpdateCh <- newStates
		case slave.Stuck:
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
		fmt.Println("ST: Sent Ack")
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
