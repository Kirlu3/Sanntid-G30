package Slave

import (
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

const ID = 1

func Slave() {
	//initialize channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	t_start := make(chan int)
	outgoing := make(chan EventMessage)
	incoming := make(chan EventMessage)

	//initialize timer
	var t_end *time.Timer = time.NewTimer(0)
	<-t_end.C
	go timer(t_start, t_end)

	//initialize sensors
	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	//initialize network
	addr := "localhost"                                //needs to be changed to master's IP
	go sender(addr+strconv.Itoa(20000+ID), outgoing)   //IP of master:20000+ID for sending to master
	go receiver(addr+strconv.Itoa(30000+ID), incoming) //IP of master:30000+ID for reveiving from master

	//initialize elevator
	var elevator Elevator
	n_elevator := fsm_onInit(elevator)
	if validElevator(n_elevator) {
		activateIO(n_elevator, elevator, t_start)
		elevator = n_elevator
	}

	//main loop (too long?)
	for {
		select {
		case msg := <-incoming: //incoming message from master
			switch msg.Event {
			case Button: //Case: new queue request
				n_elevator = fsm_onRequestButtonPress(msg.Btn, elevator) //create a new elevator struct
				if validElevator(n_elevator) {                           //if the new elevator is valid
					activateIO(n_elevator, elevator, t_start) //activate IO
					elevator = n_elevator                     //update elevator
				}
			case Light:
				elevio.SetButtonLamp(msg.Btn.Button, msg.Btn.Floor, msg.Check) //update light
			default:
				//shouldn't be any other messages sent to the slave
				continue
			}

		case btn := <-drv_buttons: //button press
			outgoing <- EventMessage{elevator, Button, btn, false} //send message to master

		case floor := <-drv_floors:
			n_elevator = fsm_onFloorArrival(floor, elevator) //create a new elevator struct
			if validElevator(n_elevator) {                   //check if the new elevator is valid
				if n_elevator.Stuck != elevator.Stuck { //if stuck status has changed
					outgoing <- EventMessage{n_elevator, Stuck, elevio.ButtonEvent{}, n_elevator.Stuck} //send message to master
				}
				activateIO(n_elevator, elevator, t_start)                                     //activate IO
				elevator = n_elevator                                                         //update elevator
				outgoing <- EventMessage{elevator, FloorArrival, elevio.ButtonEvent{}, false} //send message to master
			}
		case obs := <-drv_obstr:
			n_elevator = fsm_onObstruction(obs, elevator)
			if validElevator(n_elevator) {
				elevator = n_elevator
				outgoing <- EventMessage{elevator, Stuck, elevio.ButtonEvent{}, obs}
			}

		case <-drv_stop:
			fsm_onStopButtonPress()

		case <-t_end.C:
			n_elevator = fsm_onTimerEnd(elevator)
			if validElevator(n_elevator) {
				if n_elevator.Stuck != elevator.Stuck { //if stuck status has changed
					outgoing <- EventMessage{n_elevator, Stuck, elevio.ButtonEvent{}, n_elevator.Stuck} //send message to master
				}
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
			}
		}
	}
}

/*TODO:
- Way to check if the elevator is stuck*/

/*Things to consider:
-Is it OK to potentially not accept a new request if the request is invalid?
-Test the stuck system, I was tired when I implemented it*/
