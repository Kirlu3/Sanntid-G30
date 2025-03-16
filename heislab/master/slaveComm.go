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

/*
Listens to button presses from the slaves
Acknowledges the button press
Sends an update to the assigner for the button press
*/
func receiveButtonPress(
	callsUpdateCh chan<- struct {
		Calls   Calls
		AddCall bool
	},
	slaveToMasterOfflineButton <-chan slave.ButtonMessage,
) {
	//make a channel to receive buttons
	rx := make(chan slave.ButtonMessage)
	go bcast.Receiver(config.SlaveBasePort, rx)
	//ack channel to send acknowledgments
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+10, ack)

	var msgID []int
	for {
		select {
		case newBtn := <-rx:
			println("ST: Received button press")
			ack <- newBtn.MsgID
			fmt.Println("ST: Sent Ack", newBtn.MsgID)
			if !slices.Contains(msgID, newBtn.MsgID) {
				msgID = append(msgID, newBtn.MsgID)
				// if we've stored too many IDs, remove the oldest one. 20 is a completely arbitrary number, but leaves room for ~7 messages per slave
				if len(msgID) > 20 {
					msgID = msgID[1:]
				}
				callsUpdateCh <- makeAddCallsUpdate(newBtn)
			}
		case newBtn := <-slaveToMasterOfflineButton:
			println("ST: Received button press offline")
			callsUpdateCh <- makeAddCallsUpdate(newBtn)
		}
	}
}

/*
Listens to UDP broadcasts from the slaves
*/
func receiveElevatorUpdate(
	elevatorUpdateCh chan<- slave.Elevator,
	callsUpdateCh chan<- struct {
		Calls   Calls
		AddCall bool
	},
	slaveToMasterOfflineElevator <-chan slave.Elevator,

) {
	rx := make(chan slave.Elevator)
	elevators := [config.N_ELEVATORS]slave.Elevator{}
	go bcast.Receiver(config.SlaveBasePort+5, rx)

	for {
		select {
		case elevatorUpdate := <-rx:
			if elevators[elevatorUpdate.ID] != elevatorUpdate {
				elevators[elevatorUpdate.ID] = elevatorUpdate
				elevatorUpdateCh <- elevatorUpdate
				if elevatorUpdate.Behaviour == slave.EB_DoorOpen {
					callsUpdateCh <- makeRemoveCallsUpdate(elevatorUpdate)
				}
			}
		case elevatorUpdate := <-slaveToMasterOfflineElevator:
			elevators[elevatorUpdate.ID] = elevatorUpdate
			elevatorUpdateCh <- elevatorUpdate
			if elevatorUpdate.Behaviour == slave.EB_DoorOpen {
				callsUpdateCh <- makeRemoveCallsUpdate(elevatorUpdate)
			}
		}
	}
}

/*
Input: Elevator struct

Output: UpdateCalls struct

Function transforms the input to the right output type that is used for handling updates in the calls.
*/
func makeRemoveCallsUpdate(elevator slave.Elevator) UpdateCalls {
	var callsUpdate UpdateCalls
	callsUpdate.AddCall = false

	callsUpdate.Calls.CabCalls[elevator.ID][elevator.Floor] = true
	switch elevator.Direction {
	case slave.D_Down:
		callsUpdate.Calls.HallCalls[elevator.Floor][elevio.BT_HallDown] = true
	case slave.D_Up:
		callsUpdate.Calls.HallCalls[elevator.Floor][elevio.BT_HallUp] = true
	default:
	}
	return callsUpdate
}

/*
Input: EventMessage

Output: UpdateCalls

Function transforms the input to the right output type that is used for handling updates in the calls.
*/
func makeAddCallsUpdate(newBtn slave.ButtonMessage) UpdateCalls {
	var callsUpdate UpdateCalls
	callsUpdate.AddCall = true
	if newBtn.BtnPress.Button == elevio.BT_Cab {
		callsUpdate.Calls.CabCalls[newBtn.ElevID][newBtn.BtnPress.Floor] = true
	} else {
		callsUpdate.Calls.HallCalls[newBtn.BtnPress.Floor][newBtn.BtnPress.Button] = true
	}
	return callsUpdate
}

/*
Handles transmitting the assignments received on the toSlaveCh channel to the slaves on the SlaveBasePort-1 port.
*/
func sendMessagesToSlaves(toSlaveCh chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	masterToSlaveOfflineCh chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool) {
	tx := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	go bcast.Transmitter(config.SlaveBasePort-1, tx)

	var msg [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	for {
		select {
		case msg = <-toSlaveCh:
			fmt.Println("ST: New orders sent")
			fmt.Println(msg)
			tx <- msg
			masterToSlaveOfflineCh <- msg
		default:
			tx <- msg
		}
		time.Sleep(time.Millisecond * time.Duration(config.BroadcastMessagePeriodMs))
	}
}
