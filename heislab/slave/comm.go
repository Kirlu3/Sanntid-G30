package slave

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

type EventType int

const (
	Button       EventType = iota //In case of a button press
	FloorArrival                  //In case of a floor arrival
	Stuck                         //In case of an update to the elevator's stuck state
)

type EventMessage struct {
	MsgID    int                //Sends a unique ID for the message
	Elevator Elevator           //Sends its own elevator struct
	Event    EventType          //Sends the type of event
	Btn      elevio.ButtonEvent //Sends a button in case of a Button event
}

/*
	Transmits messages to the master

Input: The channel to receive messages that should be sent, the ID of the elevator
*/
func comm_sender(outgoing <-chan EventMessage, ID int) {
	tx := make(chan EventMessage)
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+ID, tx)
	go bcast.Receiver(config.SlaveBasePort+10+ID, ack)
	ackTimeout := make(chan int, 10)
	var needAck []EventMessage
	var out EventMessage

	for {
		select {
		case out = <-outgoing:
			fmt.Println("STx: Sending Message")
			msgID := rand.Int() //gives the message a random ID
			out.MsgID = msgID
			tx <- out
			needAck = append(needAck, out)
			ackTimeout <- msgID

		case ackID := <-ack:
			var ackIndex int
			for i := range len(needAck) {
				if needAck[i].MsgID == ackID {
					ackIndex = i
					fmt.Println("STx: Received ack")
				}
			}
			needAck[ackIndex] = needAck[len(needAck)-1] // this line gave me panic: runtime error: index out of range [-1], how is that possible? ahh length was 0 i assume
			needAck = needAck[:len(needAck)-1]

		case msgID := <-ackTimeout:
			// fmt.Println("STx: Waiting for ack")
			fmt.Println("STx: Starting timer")
			//Potential for race condition on needAck
			time.AfterFunc(time.Millisecond*time.Duration(config.ResendPeriodMs), func() {
				fmt.Println("STx: Ack timeout")
				for i := range len(needAck) {
					if needAck[i].MsgID == msgID {
						tx <- needAck[i]
						ackTimeout <- msgID
						break
					}
				}
			})
		}
	}
}

/*
	Receives messages from the master

Input: The channels to send orders and lights to the elevator, the ID of the elevator
*/
func comm_receiver(ordersRx chan<- [config.N_FLOORS][config.N_BUTTONS]bool, lightsRx chan<- [config.N_FLOORS][config.N_BUTTONS]bool, ID int) {

	rx := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	go bcast.Receiver(config.SlaveBasePort-1, rx)

	var prevMsg [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	for msg := range rx {
		if msg != prevMsg {
			fmt.Println("SRx: Received New Message")
			prevMsg = msg
			ordersRx <- msg[ID]
			//I assume there's an easier way to do this, but I need to loop through to get all active orders before sending out
			lights := [config.N_FLOORS][config.N_BUTTONS]bool{}

			for id := range config.N_ELEVATORS {
				for floor := range config.N_FLOORS {
					lights[floor][elevio.BT_Cab] = msg[ID][floor][elevio.BT_Cab]
					lights[floor][elevio.BT_HallUp] = lights[floor][elevio.BT_HallUp] || msg[id][floor][elevio.BT_HallUp]
					lights[floor][elevio.BT_HallDown] = lights[floor][elevio.BT_HallDown] || msg[id][floor][elevio.BT_HallDown]
				}
			}
			lightsRx <- lights

		} else {
			continue
		}
	}
}
