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
buttonPressRx receives button presses from the slaves, acknowledges them and sends a callsUpdate to the backupCoordinator, which forwards this information to the backups and ensures reassignments of the current calls.
*/
func buttonPressRx(
	callsUpdateChan chan<- struct {
		Calls   Calls
		AddCall bool
	},
	offlineSlaveBtnToMasterChan <-chan slave.ButtonMessage,
) {
	//channel to receive button presses
	btnPressRxChan := make(chan slave.ButtonMessage)
	go bcast.Receiver(config.SlaveButtonPort, btnPressRxChan)

	//ack channel to send acknowledgments
	ackTxChan := make(chan int)
	go bcast.Transmitter(config.SlaveAckPort, ackTxChan)

	var msgIDs []int
	for {
		select {
		case newBtn := <-btnPressRxChan:
			ackTxChan <- newBtn.MsgID
			if !slices.Contains(msgIDs, newBtn.MsgID) {
				msgIDs = append(msgIDs, newBtn.MsgID)
				// remove the oldest messageID if we've stored too many. 20 is a completely arbitrary number, but leaves room for ~7 messages per slave
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
slaveStateUpdateRx listens to UDP broadcasts from the slaves, which contains their state.
Dependent on the state of the slave, the routine sends a removeCallsUpdate to the backupCoordinator, which forwards this information to the backups and ensures reassignments of the current calls.
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
	elevators := [config.N_ELEVATORS]slave.Elevator{}
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
		if slaveStateUpdate.Behaviour == slave.EB_DoorOpen && !slaveStateUpdate.Stuck {
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
	case slave.D_Down:
		callsUpdate.Calls.HallCalls[elevator.Floor][elevio.BT_HallDown] = true
	case slave.D_Up:
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
func callsToSlavesTx(callsToSlaveChan chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	offlineCallsToSlaveChan chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool) {

	callsToSlavesTxChan := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
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
