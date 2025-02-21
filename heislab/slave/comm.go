package slave

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

type SlaveMessage struct {
	Elevator Elevator
	PrevBtn  elevio.ButtonEvent
}

func sender(elevatorTx chan Elevator, btnTx chan elevio.ButtonEvent) {
	tx := make(chan SlaveMessage)
	go bcast.Transmitter(config.SlaveBasePort+ID, tx)
	var msg SlaveMessage
	for {
		select {
		case newBtn := <-btnTx:
			msg.PrevBtn = newBtn
			tx <- msg
		case newElevator := <-elevatorTx:
			msg.Elevator = newElevator
			tx <- msg
		default:
			tx <- msg
		}
	}
}

func receiver(ordersRx chan [config.N_FLOORS][config.N_BUTTONS]bool, lightsRx chan [config.N_FLOORS][config.N_BUTTONS]bool) {

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
