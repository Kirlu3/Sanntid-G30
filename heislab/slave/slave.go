package Slave

import (
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

func Slave() {
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	t_start := make(chan bool)
	//timer init
	var t_end *time.Timer = time.NewTimer(0)
	<-t_end.C

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go timer(t_start, t_end)

	for {
		select {
		case a := <-drv_buttons:
			fsm_onRequestButtonPress(a, t_start)

		case a := <-drv_floors:
			fsm_onFloorArrival(a, t_start)

		case a := <-drv_obstr:
			fsm_onObstruction(a)

		case <-drv_stop:
			fsm_onStopButtonPress()

		case <-t_end.C:
			fsm_onTimerEnd(t_start)
		}
	}
}
