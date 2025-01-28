package fsm 

import "fmt"
import "../driver-go/elevio"
import "time"

var doorOpenDuration int := 3
var ob = make(chan bool)

type ElevatorBehaviour int
const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor int
	direction elevio.Direction
	requests var int
	behaviour ElevatorBehaviour
}

Elevator elevator := 


func timer(t_start chan bool, t_end chan bool) {
	t_on := false

	for {
		if <- t_start {
			t := time.Now() 
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
	switch Elevator.ElevatorBehaviour{
	case ElevatorBehaviour.EB_DoorOpen:
		if Elevator.floor == buttonEvent.Floor {
			t_start <- true
		} else {
			//Add to queue?
			//Don't know if we ever should get here
		}
		//Do nothing
	case ElevatorBehaviour.EB_Moving:
		Elevator.request = buttonEvent.Floor
		//There will be a bug here if we allow more than one item in the queue at a time
	case ElevatorBehaviour.EB_Idle:
		Elevator.request = buttonEvent.Floor
		switch Elevator.floor {
		case buttonEvent.Floor == Elevator.floor:
			//Open door
			t_start <- true
			Elevator.behaviour = ElevatorBehaviour.EB_DoorOpen
			elevio.SetDoorOpenLamp(true)
		case buttonEvent.Floor > Elevator.floor:
			//move up
			Elevator.direction = elevio.MD_Up
			Elevator.behaviour = ElevatorBehaviour.EB_Moving
		case ButtonEvent.Floor < Elevator.floor:
			//move down
			Elevator.direction = elevio.MD_Down
			Elevator.behaviour = ElevatorBehaviour.EB_Moving
	}
	elevio.SetMotorDirection(Elevator.direction)
	return
}}


func onFloorArrival(floor int, t_start chan bool) {
	fmt.Println("onFloorArrival")
	Elevator.floor = floor
	elevio.SetFloorIndicator(floor)

	if Elevator.request == Elevator.floor {
		Elevator.behaviour = ElevatorBehaviour.EB_DoorOpen
		Elevator.direction = elevio.MD_Stop
		elevio.SetMotorDirection(Elevator.direction)
		elevio.SetDoorOpenLamp(true)
		t_start <- true 
		//Clears queue
		Elevator.request = -1
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


func onTimerEnd() {
	//Temporary, not implemented obstruction
	for <-ob {
		ob <- true
	} 

	ob <- false
	//Checks where the next request is and sets associated direction and behaviour
	switch Elevator.request {
	case Elevator.request == -1:
		Elevator.behaviour = ElevatorBehaviour.EB_Idle	
	case Elevator.request > Elevator.floor:
		Elevator.direction = elevio.MD_Up
		Elevator.behaviour = ElevatorBehaviour.EB_Moving
	case Elevator.request < Elevator.floor:
		Elevator.direction = elevio.MD_Down
		Elevator.behaviour = ElevatorBehaviour.EB_Moving
	case Elevator.request == Elevator.floor:
		Elevator.behaviour = ElevatorBehaviour.EB_DoorOpen
	}

	//Sets the elevator to the new behaviour
	switch ElevatorBehaviour {
	case ElevatorBehaviour.EB_DoorOpen:
		t_start <- true
	case ElevatorBehaviour.EB_Moving:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(Elevator.direction)
	case ElevatorBehaviour.EB_Idle:
		elevio.SetDoorOpenLamp(false)		
	return
	}
}