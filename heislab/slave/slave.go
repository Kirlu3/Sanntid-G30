package slave

import (
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
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

	elevatorTx := make(chan Elevator)
	btnTx := make(chan elevio.ButtonEvent)
	ordersRx := make(chan [config.N_FLOORS][config.N_BUTTONS]bool)
	lightsRx := make(chan [config.N_FLOORS][config.N_BUTTONS]bool)

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
	go sender(elevatorTx, btnTx)    //Routine for sending messages to master
	go receiver(ordersRx, lightsRx) //Routine for receiving messages from master

	//initialize elevator
	var elevator Elevator
	elevator.ID = ID
	n_elevator := fsm_onInit(elevator)
	if validElevator(n_elevator) {
		activateIO(n_elevator, elevator, t_start)
		elevator = n_elevator
	}

	//send initial state to master
	//main loop (too long?)
	for {
		select {
		case msg := <-ordersRx:
			elevator.Requests = msg
			n_elevator = fsm_onRequests(elevator)
			if validElevator(n_elevator) {
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
				elevatorTx <- elevator
			}
		case msg := <-lightsRx:
			updateLights(msg)

		case btn := <-drv_buttons: //button press
			btnTx <- btn //send button to master

		case floor := <-drv_floors:
			n_elevator = fsm_onFloorArrival(floor, elevator) //create a new elevator struct
			if validElevator(n_elevator) {                   //check if the new elevator is valid
				activateIO(n_elevator, elevator, t_start) //activate IO
				elevator = n_elevator                     //update elevator
				elevatorTx <- elevator                    //send message to master
			}
		case obs := <-drv_obstr:
			n_elevator = fsm_onObstruction(obs, elevator)
			if validElevator(n_elevator) {
				elevator = n_elevator
				elevatorTx <- elevator
			}

		case <-drv_stop:
			fsm_onStopButtonPress()

		case <-t_end.C:
			n_elevator = fsm_onTimerEnd(elevator)
			if validElevator(n_elevator) {
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
				elevatorTx <- elevator
			}
		}
	}
}

/*TODO:
- Way to check if the elevator is stuck*/

/*Things to consider:
-Is it OK to potentially not accept a new request if the request is invalid?
-Test the stuck system, I was tired when I implemented it*/
