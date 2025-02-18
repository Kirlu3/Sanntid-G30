package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func receiveMessagesFromSlaves(slaveUpdate chan slave.EventMessage) {
	// receive a message from slave and put it on slaveUpdate
	var message slave.EventMessage // TODO
	slaveUpdate <- message
}

func sendMessagesToSlaves(orderAssignments chan [][]int, lightsToSlave chan slave.Calls) {
	for {
		select {
		case a := <-orderAssignments:
			fmt.Printf("a: %v\n", a)
			// for each elevator: send the orders that it has been assigned
		case s := <-lightsToSlave:
			fmt.Printf("s: %v\n", s)
			// send the lights to all the elevators (all lights or only updates? i dont think we have a good way of checking diff because who actually knows what lights are on?
			// so maybe just send all lights for each elevator)
		}

	}
}
