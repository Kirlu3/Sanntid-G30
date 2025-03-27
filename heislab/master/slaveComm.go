package master

import (
	"slices"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

/*
# Listens to buttonpresses from the slaves and sends them to be backed up. Also sends acknowledgements to messages received over the network.

Inputs: callsUpdateChan, offlineSlaveBtnToMasterChan

callsUpdateChan: sends updates about the active calls to be backed up

offlineSlaveBtnToMasterChan: receives button presses from the slave running on the same machine as the master
*/
func buttonPressRx(
	callsUpdateChan chan<- struct {
		Calls   Calls
		AddCall bool
	},
	offlineSlaveBtnToMasterChan <-chan slave.ButtonMessage,
) {
	btnPressRxChan := make(chan slave.ButtonMessage)
	go bcast.Receiver(config.SlaveButtonPort, btnPressRxChan)

	ackTxChan := make(chan int)
	go bcast.Transmitter(config.SlaveAckPort, ackTxChan)

	var msgIDs []int
	for {
		select {
		case newBtn := <-btnPressRxChan:
			ackTxChan <- newBtn.MsgID
			if !slices.Contains(msgIDs, newBtn.MsgID) {
				msgIDs = append(msgIDs, newBtn.MsgID)
				//removes the oldest message ID if the slice is longer than 20. 20 is an arbitrary number.
				if len(msgIDs) > 20 {
					msgIDs = msgIDs[1:]
				}
				callsUpdateChan <- makeAddCallsUpdate(newBtn)
			}
		case newBtn := <-offlineSlaveBtnToMasterChan:
			callsUpdateChan <- makeAddCallsUpdate(newBtn)
		}
	}
}

/*
# Listens to updates to the slaves elevator state and sends them to the assigner. Also sends updates on cleared orders to be backed up

Inputs: slaveStateUpdateChan, callsUpdateChan, offlineSlaveStateToMasterChan

slaveStateUpdateChan: sends updates about the slaves elevator state to the assigner

callsUpdateChan: sends updates about the active calls to be backed up

offlineSlaveStateToMasterChan: receives updates about the slaves elevator state from the slave running on the same machine as the master
*/
func slaveStateUpdateRx(
	slaveStateUpdateChan chan<- slave.Elevator,
	callsUpdateChan chan<- struct {
		Calls   Calls
		AddCall bool
	},
	offlineSlaveStateToMasterChan <-chan slave.Elevator,

) {
	slaveStateUpdateRxChan := make(chan slave.Elevator)
	elevators := [config.NumElevators]slave.Elevator{}
	go bcast.Receiver(config.SlaveBroadcastPort, slaveStateUpdateRxChan)

	var slaveStateUpdate slave.Elevator
	for {
		select {
		case slaveStateUpdate = <-slaveStateUpdateRxChan:
		case slaveStateUpdate = <-offlineSlaveStateToMasterChan:
		}

		if elevators[slaveStateUpdate.ID] != slaveStateUpdate {
			elevators[slaveStateUpdate.ID] = slaveStateUpdate
			slaveStateUpdateChan <- slaveStateUpdate
		}
		if slaveStateUpdate.Behaviour == slave.BehaviourDoorOpen && !slaveStateUpdate.Stuck {
			callsUpdateChan <- makeRemoveCallsUpdate(slaveStateUpdate)
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
	case slave.DirectionDown:
		callsUpdate.Calls.HallCalls[elevator.Floor][elevio.BT_HallDown] = true
	case slave.DirectionUp:
		callsUpdate.Calls.HallCalls[elevator.Floor][elevio.BT_HallUp] = true
	default:
	}
	return callsUpdate
}

/*
Input: ButtonMessage

Output: UpdateCalls

Function transforms the input to the right output type that is used for handling updates of the calls.
*/
func makeAddCallsUpdate(btnMessage slave.ButtonMessage) UpdateCalls {
	var callsUpdate UpdateCalls
	callsUpdate.AddCall = true
	if btnMessage.BtnPress.Button == elevio.BT_Cab {
		callsUpdate.Calls.CabCalls[btnMessage.ElevID][btnMessage.BtnPress.Floor] = true
	} else {
		callsUpdate.Calls.HallCalls[btnMessage.BtnPress.Floor][btnMessage.BtnPress.Button] = true
	}
	return callsUpdate
}

/*
callsToSlaveTx transmitts the calls received on callsToSlaveChan to the slaves on the SlaveCallsPort port.
*/
func callsToSlavesTx(callsToSlaveChan chan [config.NumElevators][config.NumFloors][config.NumBtns]bool,
	offlineCallsToSlaveChan chan<- [config.NumElevators][config.NumFloors][config.NumBtns]bool,
) {
	callsToSlavesTxChan := make(chan [config.NumElevators][config.NumFloors][config.NumBtns]bool)
	go bcast.Transmitter(config.SlaveCallsPort, callsToSlavesTxChan)

	callsToSlave := <-callsToSlaveChan
	for {
		select {
		case callsToSlave = <-callsToSlaveChan:
			callsToSlavesTxChan <- callsToSlave
			offlineCallsToSlaveChan <- callsToSlave
		case <-time.After(time.Millisecond * time.Duration(config.MasterBroadcastAssignedPeriodMs)):
			callsToSlavesTxChan <- callsToSlave
			offlineCallsToSlaveChan <- callsToSlave
		}
	}
}
