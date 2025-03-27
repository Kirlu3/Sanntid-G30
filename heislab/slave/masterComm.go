package slave

import (
	"math/rand/v2"
	"sync"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

type ButtonMessage struct {
	MsgID    int
	ElevID   int
	BtnPress elevio.ButtonEvent
}

/*
# buttonPressTx transmitts buttonpresses to the master until an aknowledgement for the message is received or a timeout occurs.

# If the master is located on the same machine, the buttonpresses are sent to the offlineSlaveBtnToMasterChan channel.

Input: drvBtnChan, offlineSlaveBtnToMasterChan, startSendingBtnOfflineChan, ID

drvBtnChan: receives buttonpresses from the driver

offlineSlaveBtnToMasterChan: sends buttonpresses to the master when the master is located on the same machine

startSendingBtnOfflineChan: receives instruction to start sending in offline mode
*/
func buttonPressTx(
	drvBtnChan <-chan elevio.ButtonEvent,
	offlineSlaveBtnToMasterChan chan<- ButtonMessage,
	startSendingBtnOfflineChan <-chan struct{},
	ID int,
) {
	SlaveBtnToMasterTxChan := make(chan ButtonMessage)
	ackRxChan := make(chan int)
	go bcast.Transmitter(config.SlaveButtonPort, SlaveBtnToMasterTxChan)
	go bcast.Receiver(config.SlaveAckPort, ackRxChan)

	ackTimeoutChan := make(chan int, 10)
	var needAck []ButtonMessage
	var outgoingMessage ButtonMessage
	var needAckMu sync.Mutex // Mutex for needAck to avoid any potential race condition

	sendUDP := true
mainLoop:
	for {
		select {
		case <-startSendingBtnOfflineChan:
			sendUDP = false
		case btnPress := <-drvBtnChan:

			if !sendUDP {
				offlineSlaveBtnToMasterChan <- ButtonMessage{0, ID, btnPress}
				continue mainLoop
			}

			msgID := rand.Int()
			outgoingMessage = ButtonMessage{msgID, ID, btnPress}
			SlaveBtnToMasterTxChan <- outgoingMessage
			needAckMu.Lock()
			needAck = append(needAck, outgoingMessage)
			needAckMu.Unlock()
			ackTimeoutChan <- msgID

			time.AfterFunc(time.Millisecond*time.Duration(config.ResendTimeoutMs), func() {
				needAckMu.Lock()
				needAck = removeMsgFromNeedAck(needAck, msgID)
				needAckMu.Unlock()
			})

		case msgID := <-ackRxChan:
			needAckMu.Lock()
			needAck = removeMsgFromNeedAck(needAck, msgID)
			needAckMu.Unlock()

		case msgID := <-ackTimeoutChan:

			time.AfterFunc(time.Millisecond*time.Duration(config.ResendPeriodMs), func() {
				needAckMu.Lock()
				for i := range len(needAck) {
					if needAck[i].MsgID == msgID {
						SlaveBtnToMasterTxChan <- needAck[i]
						ackTimeoutChan <- msgID
						break
					}
				}
				needAckMu.Unlock()
			})
		}
	}
}

/*
# Periodically transmits the elevator state to the master.

Input: slaveStateToMasterChan, offlineSlaveStateToMasterChan, startSendingStateOfflineChan

slaveStateToMasterChan: sends the elevator state to the master over UDP

offlineSlaveStateToMasterChan: sends the elevator state to the master when the master is located on the same machine

startSendingStateOfflineChan: receives instruction to start sending in offline mode
*/
func slaveStateTx(
	slaveStateToMasterChan <-chan Elevator,
	offlineSlaveStateToMasterChan chan<- Elevator,
	startSendingStateOfflineChan <-chan struct{},
) {
	slaveStateTxChan := make(chan Elevator)
	go bcast.Transmitter(config.SlaveBroadcastPort, slaveStateTxChan)
	var slaveState Elevator

	sendUDP := true
	for {
		select {
		case <-startSendingStateOfflineChan:
			sendUDP = false
		case slaveState = <-slaveStateToMasterChan:
		case <-time.After(time.Millisecond * time.Duration(config.SlaveBroadcastPeriodMs)):
		}

		if sendUDP {
			slaveStateTxChan <- slaveState
		} else {
			offlineSlaveStateToMasterChan <- slaveState
		}
	}
}

/*
callsFromMasterRx receives assigned calls from the master and forwards them to the local FSM on callsFromMasterChan.
The function also handles calls in offline mode by starting an annonumous goroutine that continuously listens for calls from the master on the offlineCallsToSlaveChan channel.
These calls are then handled in the same manner as the other assigned calls.

The function also updates the lights inside and outside the elevator.
*/
func callsFromMasterRx(
	callsFromMasterChan chan<- [config.NumFloors][config.NumBtns]bool,
	offlineCallsToSlaveChan <-chan [config.NumElevators][config.NumFloors][config.NumBtns]bool,
	ID int,
) {
	callsFromMasterRxChan := make(chan [config.NumElevators][config.NumFloors][config.NumBtns]bool)
	go bcast.Receiver(config.SlaveCallsPort, callsFromMasterRxChan)

	var prevCalls [config.NumElevators][config.NumFloors][config.NumBtns]bool
	var newCalls [config.NumElevators][config.NumFloors][config.NumBtns]bool
	listenUDP := true
	for {
		if listenUDP {
			select {
			case newCalls = <-callsFromMasterRxChan:
			case newCalls = <-offlineCallsToSlaveChan:
				listenUDP = false
			}
		} else {
			newCalls = <-offlineCallsToSlaveChan
		}

		if newCalls != prevCalls {
			prevCalls = newCalls
			callsFromMasterChan <- newCalls[ID]

			newLights := [config.NumFloors][config.NumBtns]bool{}

			//Gets all active calls for all elevators that can be displayed on the lights
			for elevator := range config.NumElevators {
				for floor := range config.NumFloors {
					newLights[floor][elevio.BT_Cab] = newCalls[ID][floor][elevio.BT_Cab]
					newLights[floor][elevio.BT_HallUp] = newLights[floor][elevio.BT_HallUp] || newCalls[elevator][floor][elevio.BT_HallUp]
					newLights[floor][elevio.BT_HallDown] = newLights[floor][elevio.BT_HallDown] || newCalls[elevator][floor][elevio.BT_HallDown]
				}
			}
			updateLights(newLights)
		}
	}
}

/*
Removes a message from the list of messages that require an acknowledgement.
*/
func removeMsgFromNeedAck(needAck []ButtonMessage, msgID int) []ButtonMessage {
	ackIndex := -1
	for i := range len(needAck) {
		if needAck[i].MsgID == msgID {
			ackIndex = i
		}
	}
	if len(needAck) == 0 || ackIndex == -1 {
		return needAck
	}
	needAck[ackIndex] = needAck[len(needAck)-1]
	needAck = needAck[:len(needAck)-1]
	return needAck
}
