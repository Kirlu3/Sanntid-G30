package fsm 

import "fmt"
import "Heislab/Driver-go/elevio"

type ElevatorBehaviour int
const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type Elevator struct {
	floor int
	direction MotorDirection
	behaviour ElevatorBehaviour
}

onRequestButtonPress(elevio.ButtonEvent) {
	fmt.Println("onRequestButtonPress")
	switch {
		case 
}