package slave

import (
	"fmt"
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
)

type ButtonMessage struct {
	MsgID    int                //Sends a unique ID for the message
	ElevID   int                //Sends the ID of the elevator
	BtnPress elevio.ButtonEvent //Sends a button press
}

/*
	Transmits messages to the master

Input: The channel to receive messages that should be sent, the ID of the elevator as well as the channel of button presses
Reasoning: The elevator sends all button presses to the master, and as a button event doesn't need to go by the FSM
*/
func network_buttonSender(drv_BtnChan <-chan elevio.ButtonEvent, slaveToMasterOfflineBtnChan chan<- ButtonMessage, ID int) {
	tx := make(chan ButtonMessage)
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort, tx)
	go bcast.Receiver(config.SlaveBasePort+10, ack)

	ackTimeout := make(chan int, 10)
	var needAck []ButtonMessage
	var out ButtonMessage
	var mu sync.Mutex //The chance this is necessary is extremely low, but it doesn't hurt

	masterUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(config.MasterUpdatePort, masterUpdateCh)
	var masterUpdate peers.PeerUpdate

mainLoop:
	for {
		select {
		case btn := <-drv_buttons:
			fmt.Println("STx: Button Pressed")
			select {
			case masterUpdate = <-masterUpdateCh:
			default:
			}
			if len(masterUpdate.Peers) == 0 || masterUpdate.Peers[0] == strconv.Itoa(ID) {
				select {
				case slaveToMasterOfflineButton <- ButtonMessage{0, ID, btn}:
					continue mainLoop
				case <-time.After(time.Millisecond * 100):
				}
			}

			fmt.Println("STx: Sending Button")
			msgID := rand.Int() //gives the message a random ID
			out = ButtonMessage{msgID, ID, btn}
			tx <- out
			mu.Lock()
			needAck = append(needAck, out)
			mu.Unlock()
			ackTimeout <- msgID

			time.AfterFunc(time.Millisecond*time.Duration(config.ResendTimeoutMs), func() {
				fmt.Println("STx: Message timeout", msgID)
				mu.Lock()
				oldLen := len(needAck)
				needAck = network_removeAck(needAck, msgID)
				if len(needAck) == oldLen {
					fmt.Println("STx: Ack previously received")
				}
				mu.Unlock()
			})

		case ackID := <-ack:
			fmt.Println("STx: Received Ack", ackID)
			mu.Lock()
			needAck = network_removeAck(needAck, ackID)
			mu.Unlock()

		case msgID := <-ackTimeout:
			// fmt.Println("STx: Waiting for ack")
			fmt.Println("STx: Starting timer")
			//Potential for race condition on needAck
			time.AfterFunc(time.Millisecond*time.Duration(config.ResendPeriodMs), func() {
				fmt.Println("STx: Ack timeout")
				mu.Lock()
				for i := range len(needAck) {
					if needAck[i].MsgID == msgID {
						fmt.Println("STx: Resending message", msgID)
						tx <- needAck[i]
						ackTimeout <- msgID
						break
					}
				}
				mu.Unlock()
			})
		}
	}
}

func network_broadcast(elevatorUpdateChan <-chan Elevator, slaveToMasterOfflineElevChan chan<- Elevator) {
	tx := make(chan Elevator)
	go bcast.Transmitter(config.SlaveBasePort+5, tx)
	var elevator Elevator

	masterUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(config.MasterUpdatePort, masterUpdateCh)
	var masterUpdate peers.PeerUpdate
mainLoop:
	for {
		select {
		case masterUpdate = <-masterUpdateCh:
		default:
		}
		select {
		case elevator = <-elevatorUpdateChan:
		default:
		}
		if len(masterUpdate.Peers) == 0 || masterUpdate.Peers[0] == strconv.Itoa(elevator.ID) {
			select {
			case slaveToMasterOfflineElevChan <- elevator:
				continue mainLoop
			case <-time.After(time.Millisecond * 100):
			}
		}
		tx <- elevator
		time.Sleep(time.Millisecond * time.Duration(config.BroadcastMessagePeriodMs))
	}
}

/*
Removes a message from the list of messages that require an acknoledgement
*/
func network_removeAck(needAck []ButtonMessage, msgID int) []ButtonMessage {
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

/*
	Go routine.
	Receives messages containging all orders and their assignments
	It sends the correct orders and lights to the local fsm and IO

Input: The channels to send orders and lights to the elevator, the ID of the elevator
*/
func network_receiver(
	orderReceiverChan chan<- [config.N_FLOORS][config.N_BUTTONS]bool,
	masterToSlaveOfflineOrderChan <-chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	ID int,
) {
	receiver := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	go bcast.Receiver(config.SlaveBasePort-1, rx)

	go func() {
		for msg := range masterToSlaveOfflineOrderChan {
			receiver <- msg
		}
	}()

	var prevMsg [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool

	for msg := range receiver {
		if msg != prevMsg {
			fmt.Println("SRx: Received New Message")
			prevMsg = msg
			orderReceiverChan <- msg[ID]
			//I assume there's an easier way to do this, but I need to loop through to get all active orders before sending out
			lights := [config.N_FLOORS][config.N_BUTTONS]bool{}

			for id := range config.N_ELEVATORS {
				for floor := range config.N_FLOORS {
					lights[floor][elevio.BT_Cab] = msg[ID][floor][elevio.BT_Cab]
					lights[floor][elevio.BT_HallUp] = lights[floor][elevio.BT_HallUp] || msg[id][floor][elevio.BT_HallUp]
					lights[floor][elevio.BT_HallDown] = lights[floor][elevio.BT_HallDown] || msg[id][floor][elevio.BT_HallDown]
				}
			}
			io_updateLights(lights)

		} else {
			continue
		}
	}
}
