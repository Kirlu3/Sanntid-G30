package slave

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

func Slave(id string) {
	ID, _ := strconv.Atoi(id)
	//initialize channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	t_start := make(chan int)

	tx := make(chan EventMessage)
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
	go sender(tx, ID)                   //Routine for sending messages to master
	go receiver(ordersRx, lightsRx, ID) //Routine for receiving messages from master

	//initialize elevator
	var elevator Elevator
	elevator.ID = ID
	n_elevator := fsm_onInit(elevator)
	updateLights(n_elevator.Requests)
	if validElevator(n_elevator) {
		activateIO(n_elevator, elevator, t_start)
		elevator = n_elevator
	}

	//send initial state to master
	//main loop (too long?)
	for {
		fmt.Println("FSM:New Loop")
		select {
		case msg := <-ordersRx:
			elevator.Requests = msg
			n_elevator = fsm_onRequests(elevator)
			if validElevator(n_elevator) {
				//If an order was cleared, master should get a message (if behavior = door?)
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
				if elevator.Behaviour == EB_DoorOpen {
					tx <- EventMessage{0, elevator, FloorArrival, elevio.ButtonEvent{}} //send message to master
				}
			}
		case msg := <-lightsRx:
			fmt.Println("Slave: Updating lights")
			updateLights(msg)
		case btn := <-drv_buttons: //button press
			fmt.Println("Slave: Button press")
			tx <- EventMessage{0, elevator, Button, btn} //send message to master
			fmt.Println("Slave: Button press sent")

		case floor := <-drv_floors:
			fmt.Println(floor)
			n_elevator = fsm_onFloorArrival(floor, elevator) //create a new elevator struct
			if validElevator(n_elevator) {                   //check if the new elevator is valid
				if n_elevator.Stuck != elevator.Stuck { //if stuck status has changed
					tx <- EventMessage{0, n_elevator, Stuck, elevio.ButtonEvent{}} //send message to master
				}
				activateIO(n_elevator, elevator, t_start) //activate IO
				fmt.Println("FSM: Floor: Activated IO")
				elevator = n_elevator                                               //update elevator
				tx <- EventMessage{0, elevator, FloorArrival, elevio.ButtonEvent{}} //send message to master
				fmt.Println("FSM: Completed floor arrival")
			}
		case obs := <-drv_obstr:
			n_elevator = fsm_onObstruction(obs, elevator)
			if validElevator(n_elevator) {
				activateIO(n_elevator, elevator, t_start)
				elevator = n_elevator
				tx <- EventMessage{0, elevator, Stuck, elevio.ButtonEvent{}}
			}

		case <-drv_stop:
			fsm_onStopButtonPress()

		case <-t_end.C:
			n_elevator = fsm_onTimerEnd(elevator)
			if validElevator(n_elevator) {
				if n_elevator.Stuck != elevator.Stuck { //if stuck status has changed
					tx <- EventMessage{0, n_elevator, Stuck, elevio.ButtonEvent{}} //send message to master
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
-Test the stuck system, I was tired when I implemented it
-Consider if the elevator should remove assignments from itself or not
-Fix so that lights can clear when the elevator gets an order on the floor it is idle on
*/
