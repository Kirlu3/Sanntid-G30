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
buttonPressTx transmitts buttonpresses to the master until an aknowledgement for the message is received.
*/
func buttonPressTx(
	drv_BtnChan <-chan elevio.ButtonEvent,
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
	var mu sync.Mutex //Removes all possibility of a race condition

	sendUDP := true
mainLoop:
	for {
		select {
		case <-startSendingBtnOfflineChan:
			sendUDP = false
		case btnPress := <-drv_BtnChan:

			if !sendUDP {
				offlineSlaveBtnToMasterChan <- ButtonMessage{0, ID, btnPress}
				continue mainLoop
			}

			msgID := rand.Int()
			outgoingMessage = ButtonMessage{msgID, ID, btnPress}
			SlaveBtnToMasterTxChan <- outgoingMessage
			mu.Lock()
			needAck = append(needAck, outgoingMessage)
			mu.Unlock()
			ackTimeoutChan <- msgID

			time.AfterFunc(time.Millisecond*time.Duration(config.ResendTimeoutMs), func() {
				mu.Lock()
				needAck = removeMsgFromNeedAck(needAck, msgID)
				mu.Unlock()
			})

		case msgID := <-ackRxChan:
			mu.Lock()
			needAck = removeMsgFromNeedAck(needAck, msgID)
			mu.Unlock()

		case msgID := <-ackTimeoutChan:

			time.AfterFunc(time.Millisecond*time.Duration(config.ResendPeriodMs), func() {
				mu.Lock()
				for i := range len(needAck) {
					if needAck[i].MsgID == msgID {
						SlaveBtnToMasterTxChan <- needAck[i]
						ackTimeoutChan <- msgID
						break
					}
				}
				mu.Unlock()
			})
		}
	}
}

/*
slaveStateTx handles periodic transmission of the elevator´s state to the master.
It reads the elevator´s state from slaveStateToMasterChan.
If the master is in offline mode, the state of the elevator is sent to the channel offlineSlaveStateToMasterChan.
Otherwise, the state is broadcasted to the master.
The function continuously checks for master updates and ensures that the elevator's state is transmitted at regular intervals.
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
	callsFromMasterChan chan<- [config.N_FLOORS][config.N_BUTTONS]bool,
	offlineCallsToSlaveChan <-chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	ID int,
) {
	callsFromMasterRxChan := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	go bcast.Receiver(config.SlaveCallsPort, callsFromMasterRxChan)

	var prevCalls [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	var newCalls [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
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

			newLights := [config.N_FLOORS][config.N_BUTTONS]bool{}

			//Gets all active calls for all elevators that can be displayed on the lights
			for elevator := range config.N_ELEVATORS {
				for floor := range config.N_FLOORS {
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
