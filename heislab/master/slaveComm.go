package master

import (
	"fmt"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func receiveMessagesFromSlaves(slaveUpdate chan<- slave.EventMessage) {
	for slaveID := 1; slaveID <= config.N_ELEVATORS; slaveID++ {
		go receiveMessageFromSlave(slaveUpdate, slaveID)
	}
}

func receiveMessageFromSlave(slaveUpdate chan<- slave.EventMessage, slaveID int) {
	//rx channel for receiving from each slave
	rx := make(chan slave.EventMessage)
	go bcast.Receiver(config.SlaveBasePort+slaveID, rx)
	//ack channel to send an acknowledgment to each slave
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+slaveID+10, ack)
	var msgID int
	for msg := range rx {
		ack <- msg.MsgID
		if msg.MsgID != msgID {
			println("Received new message")
			msgID = msg.MsgID
			slaveUpdate <- msg
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
