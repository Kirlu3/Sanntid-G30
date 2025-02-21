package slave

import (
	"math/rand/v2"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

type EventType int

const (
	Button       EventType = iota //In case of a button press or queue update, both ways
	Light                         //In case of a light update, only from master
	FloorArrival                  //In case of a floor arrival, only from slave
	Stuck                         //In case of a stuck elevator, only from slave
)

type EventMessage struct {
	MsgID    int
	Elevator Elevator           //Sends its own elevator struct, always
	Event    EventType          //Sends the type of event
	Btn      elevio.ButtonEvent //Sends a button in case of Button or Light
	Check    bool               //Sends a boolean for either Stuck or Light
}

func sender(outgoing <-chan EventMessage) {
	tx := make(chan EventMessage)
	acc := make(chan int)
	go bcast.Transmitter(config.SlaveBasePort+ID, tx)
	go bcast.Receiver(config.SlaveBasePort+10, acc)
	for {
		select {
		case out := <-outgoing:
			msgID := rand.Int() //gives the message a random ID
			out.MsgID = msgID
			tx <- out

		}
	}
}

func receiver(ordersRx chan<- [config.N_FLOORS][config.N_BUTTONS]bool, lightsRx chan<- [config.N_FLOORS][config.N_BUTTONS]bool) {

	rx := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	go bcast.Receiver(config.SlaveBasePort+ID, rx)

	var prevMsg [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	var msg [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool
	for {
		msg = <-rx
		if msg != prevMsg {
			prevMsg = msg
			ordersRx <- msg[ID-1]
			//I assume there's an easier way to do this, but I need to loop through to get all active orders before sending out
			lights := [config.N_FLOORS][config.N_BUTTONS]bool{}
			lights[0:config.N_FLOORS][elevio.BT_Cab] = msg[ID][0:config.N_FLOORS][elevio.BT_Cab]
			for i := 0; i < config.N_ELEVATORS; i++ {
				for j := 0; j < config.N_FLOORS; j++ {
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

/*TODO
- Add a message frequency to sending
- Fix the for loops?*/
