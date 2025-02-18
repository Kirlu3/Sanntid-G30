package master

import Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"

func receiveMessagesFromSlaves(slaveUpdate chan Slave.EventMessage) {
	// receive a message from slave and put it on slaveUpdate
	var message Slave.EventMessage // TODO
	slaveUpdate <- message
}
