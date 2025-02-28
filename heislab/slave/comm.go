package slave

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

type EventType int

const (
	Button       EventType = iota //In case of a button press or queue update, both ways
	FloorArrival                  //In case of a floor arrival, only from slave
	Stuck                         //In case of a stuck elevator, only from slave
)

type EventMessage struct {
	MsgID    int
	Elevator Elevator           //Sends its own elevator struct, always
	Event    EventType          //Sends the type of event
	Btn      elevio.ButtonEvent //Sends a button in case of Button or Light
}

func sender(outgoing <-chan EventMessage, ID int) {
	tx := make(chan EventMessage)
	ack := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+ID, tx)
	go bcast.Receiver(config.SlaveBasePort+10+ID, ack)
	var msgID int
	ackTimeout := make(chan int, 2)
	var needAck []int
	var timerRunning []int
	var out EventMessage

	for {
		select {
		case out = <-outgoing:
			fmt.Println("STx: Sending Message")
			msgID = rand.Int() //gives the message a random ID
			out.MsgID = msgID
			tx <- out
			needAck = append(needAck, msgID)
			ackTimeout <- msgID

		case ackID := <-ack:
			if slices.Contains(needAck, ackID) {
				needAck[slices.Index(needAck, ackID)] = needAck[len(needAck)-1]
				needAck = needAck[:len(needAck)-1]
				fmt.Println("STx: Received ack")
			}

		case msgID := <-ackTimeout:
			// fmt.Println("STx: Waiting for ack")
			if !slices.Contains(timerRunning, msgID) {
				fmt.Println("STx: Starting timer")
				timerRunning = append(timerRunning, msgID)
				time.AfterFunc(time.Millisecond*1000, func() {
					fmt.Println("STx: Ack timeout", timerRunning)
					timerRunning[slices.Index(timerRunning, msgID)] = timerRunning[len(timerRunning)-1]
					timerRunning = timerRunning[:len(timerRunning)-1]
					if slices.Contains(needAck, msgID) {
						fmt.Println("STx: No ack received")
						tx <- out
						ackTimeout <- msgID
					} else {
						fmt.Println("STx: Ack previously received")
					}
				})
			}

		}
	}
}

func receiver(ordersRx chan<- [config.N_FLOORS][config.N_BUTTONS]bool, lightsRx chan<- [config.N_FLOORS][config.N_BUTTONS]bool, ID int) {

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

			for i := range config.N_ELEVATORS {
				for j := range config.N_FLOORS {
					lights[j][elevio.BT_Cab] = msg[ID][j][elevio.BT_Cab]
					lights[j][elevio.BT_HallUp] = lights[j][elevio.BT_HallUp] || msg[i][j][elevio.BT_HallUp]
					lights[j][elevio.BT_HallDown] = lights[j][elevio.BT_HallDown] || msg[i][j][elevio.BT_HallDown]
				}
			}
			lightsRx <- lights
		} else {
			continue
		}
	}
}
