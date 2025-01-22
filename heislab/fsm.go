package fsm 

import "fmt"
import "Heislab/Driver-go/elevio"
import "time"

doorOpenDuration := 3

type ElevatorBehaviour int
const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor int
	direction MotorDirection
	request int
	behaviour ElevatorBehaviour
}

timer(t_start, t_end) {
	t_on := false
	for {
		if <-t_start {
			t := time.Now()
			t_on = true
		}
		if time.Now() - t > doorOpenDuration && t_on {
			t_end <- true
			t_on = false
		}
	}
}

onRequestButtonPress(elevio.ButtonEvent) {
	fmt.Println("onRequestButtonPress")
	switch ElevatorBehaviour{
	case ElevatorBehaviour.EB_DoorOpen:
		if Elevator.floor == ButtonEvent.Floor {
			//Open door
		} else {
			//Add to queue?
			//Don't know if we ever should get here
		}
		//Do nothing
	case ElevatorBehaviour.EB_Moving:
		Elevator.request = ButtonEvent.Floor
		//There will be a bug here if we allow more than one item in the queue at a time
	case ElevatorBehaviour.EB_Idle:
		Elevator.request = ButtonEvent.Floor
		switch Elevator.floor {
			case ButtonEvent.Floor == Elevator.floor:
				//Open door
			case ButtonEvent.Floor > Elevator.floor:
				Elevator.direction = elevio.MD_Up
				Elevator.behaviour = ElevatorBehaviour.EB_Moving
			case ButtonEvent.Floor < Elevator.floor:
				Elevator.direction = elevio.MD_Down
				Elevator.behaviour = ElevatorBehaviour.EB_Moving
		}
		elevio.SetMotorDirection(Elevator.direction)
}