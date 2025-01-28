package request

import "fmt"
import "../driver-go/elevio"
import "../fsm"

type DirectionBehaviorPair int
const (
	direction elevio.Direction
	behavior fsm.ElevatorBehaviour
)

func requests_above(elevator fsm.Elevator) {
	for f := elevator.floor ; i < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.request[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(elevator fsm.Elevator) {
	for f := 0 ; i < elevator.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if elevator.request[f][btn] {
				return true
			}
		}
	}
	return false
}

func request_here(elevator fsm.Elevator) {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if elevator.request[elevator.floor][btn] {
			return true
		}
	return false
}

func requests_chooseDirection(elevator fsm.Elevator) {
	switch elevator.direction {
	case MD_Up:
		if requests_above(elevator) {
			return {elevio.Direction.D_Up, fsm.ElevatorBehaviour.EB_Moving} DirectionBehaviorPair
		}
		if request_here(elevator) {
			return {elevio.Direction.D_Down, fsm.ElevatorBehaviour.EB_DoorOpen} DirectionBehaviorPair
		}
		if requests_below(elevator) {
			return {elevio.Direction.D_Down, fsm.ElevatorBehaviour.EB_Moving} DirectionBehaviorPair
		}
		else {
			return {elevio.Direction.D_Stop, fsm.ElevatorBehaviour.EB_Idle} DirectionBehaviorPair
		}
	case MD_Down:
		if requests_below(elevator) {
			return {elevio.Direction.D_Down, fsm.ElevatorBehaviour.EB_Moving} DirectionBehaviorPair
		}
		if request_here(elevator) {
			return {elevio.Direction.D_Up, fsm.ElevatorBehaviour.EB_DoorOpen} DirectionBehaviorPair
		}
		if requests_above(elevator) {
			return {elevio.Direction.D_Up, fsm.ElevatorBehaviour.EB_Moving} DirectionBehaviorPair
		}
		else {
			return {elevio.Direction.D_Stop, fsm.ElevatorBehaviour.EB_Idle} DirectionBehaviorPair
		}
	case MD_Stop:
		if request_here(elevator) {
			return {elevio.Direction.D_Stop, fsm.ElevatorBehaviour.EB_DoorOpen} DirectionBehaviorPair
		}
			if requests_above(elevator) {
			return {elevio.Direction.D_Up, fsm.ElevatorBehaviour.EB_Moving} DirectionBehaviorPair
		}
		if requests_below(elevator) {
			return {elevio.Direction.D_Down, fsm.ElevatorBehaviour.EB_Moving} DirectionBehaviorPair
		}
		else {
			return {elevio.Direction.D_Stop, fsm.ElevatorBehaviour.EB_Idle} DirectionBehaviorPair
		}
	}
}