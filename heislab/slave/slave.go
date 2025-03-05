package slave

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

// Initialization and main loop of the slave module
func Slave(id string, online <-chan bool, goOnlineCalls chan<- [config.N_FLOORS][config.N_BUTTONS]bool) {
	ID, _ := strconv.Atoi(id)
	//initialize channels
	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	t_start := make(chan int)

	tx := make(chan EventMessage)
	ordersRx := make(chan [config.N_FLOORS][config.N_BUTTONS]bool, 2)
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
	go comm_sender(tx, ID)
	go comm_receiver(ordersRx, lightsRx, ID)

	//initialize elevator
	var elevator Elevator
	elevator.ID = ID
	io_updateLights(elevator.Requests)

	isOnline := true
	var activeLights [config.N_FLOORS][config.N_BUTTONS]bool

	n_elevator := fsm_onInit(elevator)
	elevator = elevator_updateElevator(n_elevator, elevator, tx, t_start, isOnline)

	//main loop
	for {
		fmt.Println("FSM:New Loop")
		select {
		case newRequests := <-ordersRx:
			fmt.Println("Slave: Updating orders")

			elevator.Requests = newRequests
			n_elevator = fsm_onRequests(elevator)
			elevator = elevator_updateElevator(n_elevator, elevator, tx, t_start, isOnline)

		case activeLights = <-lightsRx:
			fmt.Println("Slave: Updating lights")
			io_updateLights(activeLights)

		case btn := <-drv_buttons: //button press
			if !isOnline {
				newRequests := elevator.Requests
				newRequests[btn.Floor][btn.Button] = true
				ordersRx <- newRequests
				break
			}
			fmt.Println("Slave: Button press")
			tx <- EventMessage{0, elevator, Button, btn} //send message to master

		case floor := <-drv_floors:
			fmt.Println("FSM: Floor arrival", floor)
			n_elevator = fsm_onFloorArrival(floor, elevator)
			elevator = elevator_updateElevator(n_elevator, elevator, tx, t_start, isOnline)

		case obs := <-drv_obstr:
			n_elevator = fsm_onObstruction(obs, elevator)
			elevator = elevator_updateElevator(n_elevator, elevator, tx, t_start, isOnline)

		case <-drv_stop:
			fsm_onStopButtonPress()

		case <-t_end.C:
			fmt.Println("FSM: Timer end")
			n_elevator = fsm_onTimerEnd(elevator)
			elevator = elevator_updateElevator(n_elevator, elevator, tx, t_start, isOnline)

		case isOnline = <-online:
			if isOnline {
				fmt.Println("Slave: Going online")
				goOnlineCalls <- elevator.Requests
				break
			}
			fmt.Println("Slave: Going offline")
			ordersRx <- activeLights
		}

		if !isOnline {
			io_updateLights(elevator.Requests)
		}
	}
}
