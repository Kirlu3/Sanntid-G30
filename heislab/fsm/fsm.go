package fsm 

import "fmt"
import "../driver-go/elevio"
import "time"

var doorOpenDuration = time.Second * 3

var ob = make(chan bool)

type ElevatorBehaviour int
const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor int 
	direction elevio.MotorDirection 
	requests [0]int
	behaviour ElevatorBehaviour
}
var elevator = Elevator{-1, elevio.MD_Stop, [0]int{}, EB_Idle}

func timer(t_start chan bool, t_end chan bool) {
	t_on := false
	t := time.Now() 

	for {
		if <- t_start {
			t = time.Now() 
			t_on = true
		}
		if time.Now() - t > doorOpenDuration && t_on {
			t_end <- true
			t_on = false
		}
	}
}

func onRequestButtonPress(buttonEvent elevio.ButtonEvent, t_start chan bool) {
	fmt.Println("onRequestButtonPress")
	switch elevator.behaviour{
	case EB_DoorOpen:
		if elevator.floor == buttonEvent.Floor {
			t_start <- true
		} else {
			//Add to queue?
			//Don't know if we ever should get here
		}
		//Do nothing
	case EB_Moving:
		elevator.requests = buttonEvent.Floor
		//There will be a bug here if we allow more than one item in the queue at a time
	case EB_Idle:
		elevator.requests = buttonEvent.Floor
		switch elevator.floor {
		case buttonEvent.Floor == elevator.floor:
			//Open door
			t_start <- true
			elevator.behaviour = EB_DoorOpen
			elevio.SetDoorOpenLamp(true)
		case buttonEvent.Floor > elevator.floor:
			//move up
			elevator.direction = elevio.MD_Up
			elevator.behaviour = EB_Moving
		case ButtonEvent.Floor < elevator.floor:
			//move down
			elevator.direction = elevio.MD_Down
			elevator.behaviour = EB_Moving
	}
	elevio.SetMotorDirection(elevator.direction)
	return
}}


func onFloorArrival(floor int, t_start chan bool) {
	fmt.Println("onFloorArrival")
	elevator.floor = floor
	elevio.SetFloorIndicator(floor)

	if elevator.requests == elevator.floor {
		elevator.behaviour = EB_DoorOpen
		elevator.direction = elevio.MD_Stop
		elevio.SetMotorDirection(elevator.direction)
		elevio.SetDoorOpenLamp(true)
		t_start <- true 
		//Clears queue
		elevator.requests = -1
	}
	//Send a completion message?
	return
}

//not implemented yet? This is an attempt that might easily make bugs
func onObstruction(obstruction bool) {
	<-ob
	ob <- obstruction 
	return
}

func onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
	return
}
time.Time(time.Now().Year(), 01, 01, 0, 0, 0, 0, time.UTC)tion
	for <-ob {
		ob <- true
	} 

	ob <- false
	//Checks where the next request is and sets associated direction and behaviour
	switch elevator.requests {
	case -1:
		elevator.behaviour = EB_Idle	
	case elevator.requests > elevator.floor:
		elevator.direction = elevio.MD_Up
		elevator.behaviour = EB_Moving
	case elevator.requests < elevator.floor:
		elevator.direction = elevio.MD_Down
		elevator.behaviour = EB_Moving
	case elevator.requests == elevator.floor:
		elevator.behaviour = EB_DoorOpen
	}

	//Sets the elevator to the new behaviour
	switch ElevatorBehaviour {
	case EB_DoorOpen:
		t_start <- true
	case EB_Moving:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevator.direction)
	case EB_Idle:
		elevio.SetDoorOpenLamp(false)		
	return
	}
}