package Slave

import (
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

const id = 1

func Slave() {
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	t_start := make(chan bool)
	outgoing := make(chan EventMessage)
	incoming := make(chan EventMessage)

	//timer init
	var t_end *time.Timer = time.NewTimer(0)
	<-t_end.C

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer(t_start, t_end)

	addr := "localhost"
	go sender(addr, outgoing)
	go receiver(addr, incoming)

	var elevator Elevator

	n_elevator := fsm_onInit(elevator)
	if validElevator(n_elevator) {
		activateIO(n_elevator, elevator, t_start)
		elevator = n_elevator
	}

	for {
		select {
		case a := <-drv_buttons:
			n_elevator = fsm_onRequestButtonPress(a, elevator)
			if validElevator(n_elevator) {
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
			}
		case a := <-drv_floors:
			n_elevator = fsm_onFloorArrival(a, elevator)
			if validElevator(n_elevator) {
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
			}
		case a := <-drv_obstr:
			fsm_onObstruction(a)

		case <-drv_stop:
			fsm_onStopButtonPress()

		case <-t_end.C:
			n_elevator = fsm_onTimerEnd(elevator)
			if validElevator(n_elevator) {
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
			}
		}
	}
}
