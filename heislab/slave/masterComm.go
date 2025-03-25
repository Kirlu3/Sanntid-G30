package slave

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/alive"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

type ButtonMessage struct {
	MsgID    int                //Sends a unique ID for the message
	ElevID   int                //Sends the ID of the elevator
	BtnPress elevio.ButtonEvent //Sends a button press
}

/*
buttonPressTx transmitts buttonpresses to the master until an aknowledgement for the message is received.
*/
func buttonPressTx(drv_BtnChan <-chan elevio.ButtonEvent, offlineSlaveBtnToMasterChan chan<- ButtonMessage, ID int) {
	SlaveBtnToMasterTxChan := make(chan ButtonMessage)
	ackRxChan := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort, SlaveBtnToMasterTxChan)
	go bcast.Receiver(config.SlaveBasePort+10, ackRxChan)

	ackTimeoutChan := make(chan int, 10)
	var needAck []ButtonMessage
	var outgoingMessage ButtonMessage
	var mu sync.Mutex //Removes all possibility of a race condition

	masterUpdateRxChan := make(chan alive.AliveUpdate)
	go alive.Receiver(config.MasterUpdatePort, masterUpdateRxChan)
	var masterUpdate alive.AliveUpdate

mainLoop:
	for {
		select {
		case masterUpdate = <-masterUpdateRxChan:
			continue mainLoop
		case btnPress := <-drv_BtnChan:
			fmt.Println("STx: Button Pressed")

			if len(masterUpdate.Alive) == 0 || masterUpdate.Alive[0] == strconv.Itoa(ID) {
				select {
				case offlineSlaveBtnToMasterChan <- ButtonMessage{0, ID, btnPress}:
					continue mainLoop
				case <-time.After(time.Millisecond * 100):
				}
			}

			fmt.Println("STx: Sending Button")
			msgID := rand.Int()
			outgoingMessage = ButtonMessage{msgID, ID, btnPress}
			SlaveBtnToMasterTxChan <- outgoingMessage
			mu.Lock()
			needAck = append(needAck, outgoingMessage)
			mu.Unlock()
			ackTimeoutChan <- msgID

			time.AfterFunc(time.Millisecond*time.Duration(config.ResendTimeoutMs), func() {
				//fmt.Println("STx: Message timeout", msgID)
				mu.Lock()
				oldLen := len(needAck)
				needAck = removeMsgFromNeedAck(needAck, msgID)
				if len(needAck) == oldLen {
					//fmt.Println("STx: Ack previously received")
				}
				mu.Unlock()
			})

		case msgID := <-ackRxChan:
			fmt.Println("STx: Received Ack", msgID)
			mu.Lock()
			needAck = removeMsgFromNeedAck(needAck, msgID)
			mu.Unlock()

		case msgID := <-ackTimeoutChan:
			// fmt.Println("STx: Waiting for ack")
			//fmt.Println("STx: Starting timer")
			//Potential for race condition on needAck
			time.AfterFunc(time.Millisecond*time.Duration(config.ResendPeriodMs), func() {
				//fmt.Println("STx: Ack timeout")
				mu.Lock()
				for i := range len(needAck) {
					if needAck[i].MsgID == msgID {
						//fmt.Println("STx: Resending message", msgID)
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
func slaveStateTx(slaveStateToMasterChan <-chan Elevator, offlineSlaveStateToMasterChan chan<- Elevator) {
	slaveStateTxChan := make(chan Elevator)
	go bcast.Transmitter(config.SlaveBasePort+5, slaveStateTxChan)
	var slaveState Elevator

	masterUpdateRxChan := make(chan alive.AliveUpdate)
	go alive.Receiver(config.MasterUpdatePort, masterUpdateRxChan)
	var masterUpdate alive.AliveUpdate

mainLoop:
	for {
		select {
		case masterUpdate = <-masterUpdateRxChan:
			//Do nothing
		case slaveState = <-slaveStateToMasterChan:
			//Do nothing
		case <-time.After(time.Millisecond * time.Duration(config.SlaveBroadcastPeriodMs)):
			//Do nothing
		}

		if len(masterUpdate.Alive) == 0 || masterUpdate.Alive[0] == strconv.Itoa(slaveState.ID) {
			select {
			case offlineSlaveStateToMasterChan <- slaveState:
				continue mainLoop
			case <-time.After(time.Millisecond * 100):
			}
		}
		slaveStateTxChan <- slaveState
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
	go bcast.Receiver(config.SlaveBasePort-1, callsFromMasterRxChan)

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
			fmt.Println("SRx: Received New Message")
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
