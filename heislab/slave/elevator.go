package slave

import (
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

type ElevatorBehaviour int

const (
	BehaviourIdle ElevatorBehaviour = iota
	BehaviourDoorOpen
	BehaviourMoving
)

type ElevatorDirection int

const (
	DirectionDown ElevatorDirection = -1
	DirectionStop ElevatorDirection = 0
	DirectionUp   ElevatorDirection = 1
)

type Elevator struct {
	ID          int
	Floor       int
	Direction   ElevatorDirection
	Calls       [config.NumFloors][config.NumBtns]bool
	Behaviour   ElevatorBehaviour
	Stuck       bool
	Obstruction bool
}

/*
# Checks if an elevator has a state within the correct bounds

Input: The elevator to be checked

Returns: True if the elevator is valid, false otherwise
*/
func validElevator(elevator Elevator) bool {
	return elevator.Behaviour >= BehaviourIdle && elevator.Behaviour <= BehaviourMoving && //Behaviour in bounds
		elevator.Direction >= DirectionDown && elevator.Direction <= DirectionUp && //Direction in bounds
		elevator.Floor > -1 && elevator.Floor < config.NumFloors && //Floor in bounds
		!elevator.Calls[config.NumFloors-1][elevio.BT_HallUp] && !elevator.Calls[0][elevio.BT_HallDown] && //no impossible Calls
		!(elevator.Behaviour == BehaviourMoving && elevator.Direction == DirectionStop) //no Behaviour moving without Direction
}

/*
# Updates the state of elevator with the state of newElevator if its is valid, sends the new state to the braodcaster, and starts the correct timer.

Input: The new state of the elevator, the current state of the elevator, the channel to send the new state to the broadcaster, and the channel to send the timer duration to the timer.

Returns: The updated elevator struct
*/
func updateElevatorState(newElevator Elevator,
	elevator Elevator,
	slaveStateToMasterChan chan<- Elevator,
	stateTimer *time.Timer,
) Elevator {
	if validElevator(newElevator) {
		slaveStateToMasterChan <- newElevator

		switch newElevator.Behaviour {
		case BehaviourDoorOpen:
			if newElevator.Calls != elevator.Calls || newElevator.Obstruction {
				stateTimer.Reset(time.Second * time.Duration(config.DoorOpenDuration))
			}
		case BehaviourMoving:
			stateTimer.Reset(time.Second * time.Duration(config.TimeBetweenFloors))
		}

		return newElevator
	}
	return elevator
}

/*
Returns true if there are calls above the elevator's current floor

Else returns false
*/
func callsAboveElevator(elevator Elevator) bool {
	for f := elevator.Floor + 1; f < config.NumFloors; f++ {
		for btn := range config.NumBtns {
			if elevator.Calls[f][btn] {
				return true
			}
		}
	}
	return false
}

/*
Returns true if there are calls below the elevator's current floor

Else returns false
*/
func callsBelowElevator(elevator Elevator) bool {
	for f := range elevator.Floor {
		for btn := range config.NumBtns {
			if elevator.Calls[f][btn] {
				return true
			}
		}
	}
	return false
}

/*
Returns true if there are calls at the elevator's current floor

Else returns false
*/
func callsAtCurrentFloor(elevator Elevator) bool {
	for btn := range config.NumBtns {
		if elevator.Calls[elevator.Floor][btn] {
			return true
		}
	}
	return false
}

/*
# Chooses the direction and behaviour of the elevator based on the calls and current state of the elevator

Input: The elevator

Returns: The chosen direction and behaviour for the elevator
*/
func chooseElevatorDirection(elevator Elevator) (ElevatorDirection, ElevatorBehaviour) {
	switch elevator.Direction {
	case DirectionUp:
		if elevator.Calls[elevator.Floor][elevio.BT_HallUp] || elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return DirectionUp, BehaviourDoorOpen
		} else if callsAboveElevator(elevator) {
			return DirectionUp, BehaviourMoving
		} else if callsAtCurrentFloor(elevator) {
			return DirectionDown, BehaviourDoorOpen
		} else if callsBelowElevator(elevator) {
			return DirectionDown, BehaviourMoving
		} else {
			return DirectionUp, BehaviourIdle
		}
	case DirectionDown:
		if elevator.Calls[elevator.Floor][elevio.BT_HallDown] || elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return DirectionDown, BehaviourDoorOpen
		} else if callsBelowElevator(elevator) {
			return DirectionDown, BehaviourMoving
		} else if callsAtCurrentFloor(elevator) {
			return DirectionUp, BehaviourDoorOpen
		} else if callsAboveElevator(elevator) {
			return DirectionUp, BehaviourMoving
		} else {
			return DirectionDown, BehaviourIdle
		}
	case DirectionStop:
		if callsAtCurrentFloor(elevator) {
			return DirectionStop, BehaviourDoorOpen
		} else if callsAboveElevator(elevator) {
			return DirectionUp, BehaviourMoving
		} else if callsBelowElevator(elevator) {
			return DirectionDown, BehaviourMoving
		} else {
			return DirectionStop, BehaviourIdle
		}
	default:
		panic("Invalid elevator direction")
	}
}

/*
Returns true if the elevator should stop at the current floor

Else returns false
*/
func shouldElevatorStop(elevator Elevator) bool {
	switch elevator.Direction {
	case DirectionDown:
		if elevator.Calls[elevator.Floor][elevio.BT_HallDown] {
			return true
		}
		if elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return true
		}
		if !callsBelowElevator(elevator) {
			return true
		}
		return false
	case DirectionUp:
		if elevator.Calls[elevator.Floor][elevio.BT_HallUp] {
			return true
		}
		if elevator.Calls[elevator.Floor][elevio.BT_Cab] {
			return true
		}
		if !callsAboveElevator(elevator) {
			return true
		}
		return false
	default:
		return true
	}
}

/*
# Clears calls depending on the direction of the elevator. Will only clear hall calls in one direction at a time.

Input: the elevator

Returns: the elevator with cleared calls at the current floor, in addition to a new direction in the case of no initial direction.
*/
func clearCallsAtCurrentFloor(elevator Elevator) Elevator {
	elevator.Calls[elevator.Floor][elevio.BT_Cab] = false
	switch elevator.Direction {
	case DirectionUp:
		elevator.Calls[elevator.Floor][elevio.BT_HallUp] = false
	case DirectionDown:
		elevator.Calls[elevator.Floor][elevio.BT_HallDown] = false
	default:
		if elevator.Calls[elevator.Floor][elevio.BT_HallUp] {
			elevator.Calls[elevator.Floor][elevio.BT_HallUp] = false
			elevator.Direction = DirectionUp
		} else if elevator.Calls[elevator.Floor][elevio.BT_HallDown] {
			elevator.Calls[elevator.Floor][elevio.BT_HallDown] = false
			elevator.Direction = DirectionDown
		}
	}
	return elevator
}
