package slave

import (
	"fmt"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
)

/*
	The main finite state machine of the elevator.
	Explained in more detail in the README

Input: the elevator ID and all relevant channels
*/
func fsm(ID int,
	slaveStateToMasterChan chan<- Elevator,
	callsFromMasterChan <-chan [config.N_FLOORS][config.N_BUTTONS]bool,
	drvNewFloorChan <-chan int,
	drvObstrChan <-chan bool,
	drvStopChan <-chan bool,
	timerDurationChan chan int,
	timer *time.Timer,
) {

	var elevator Elevator
	elevator.ID = ID
	updateLights(elevator.Calls)

	newElevator := initElevator(elevator)
	elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

	for {
		activateElevatorIO(elevator)
		select {
		case newCalls := <-callsFromMasterChan:
			elevator.Calls = newCalls
			newElevator = onNewCalls(elevator)
			elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

		case floor := <-drvNewFloorChan:
			newElevator = onFloorArrival(floor, elevator)
			elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

		case obstr := <-drvObstrChan:
			newElevator = onObstruction(obstr, elevator)
			elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

		case <-drvStopChan:
			onStopButtonPress()

		case <-timer.C:
			newElevator = onTimerEnd(elevator)
			elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)
		}
	}
}

/*
	Activates when the elevator is initialized

Input: elevator object

Returns: the new elevator object with initialized direction and behaviour
*/
func initElevator(elevator Elevator) Elevator {
	elevator.Direction = D_Up
	elevator.Behaviour = EB_Moving
	return elevator
}

/*
	Activates when the elevator receives new calls

Input: the old elevator object with updated calls

Returns: the new elevator object with updated direction and behaviour
*/
func onNewCalls(elevator Elevator) Elevator {
	switch elevator.Behaviour {
	case EB_Idle:
		direction, behaviour := chooseElevatorDirection(elevator)
		elevator.Direction = direction
		elevator.Behaviour = behaviour
		if elevator.Behaviour == EB_DoorOpen {
			elevator = clearCallsAtCurrentFloor(elevator)
		}
	case EB_DoorOpen:
		if !elevator.Obstruction {
			elevator = clearCallsAtCurrentFloor(elevator)
		}
	}

	return elevator
}

/*
	Activates when the elevator floor sensor is triggered

Input: the elevator and the new floor

Returns: the elevator with the floor updated
*/
func onFloorArrival(newFloor int, elevator Elevator) Elevator {
	if !elevator.Obstruction {
		elevator.Stuck = false //if the elevator arrives at a floor, it is not stuck
	}
	elevator.Floor = newFloor
	switch elevator.Behaviour {
	case EB_Moving:
		if shouldElevatorStop(elevator) {
			if callsAtCurrentFloor(elevator) {
				newElevator := clearCallsAtCurrentFloor(elevator)
				if newElevator.Calls == elevator.Calls {
					switch elevator.Direction {
					case D_Down:
						elevator.Direction = D_Up
					case D_Up:
						elevator.Direction = D_Down
					}
				}
				elevator = clearCallsAtCurrentFloor(elevator)
				elevator.Behaviour = EB_DoorOpen
			} else {
				direction, behaviour := chooseElevatorDirection(elevator)
				elevator.Direction = direction
				elevator.Behaviour = behaviour
			}
		}
	}
	return elevator
}

/*
	Activates when the obstruction sensor is triggered

Input: the state of the obstructuin switch and the elevator object.

Returns: the elevator object with updated state depending on obstruction.
*/
func onObstruction(obstruction bool, elevator Elevator) Elevator {
	elevator.Obstruction = obstruction
	return elevator
}

/*
	Activates when the stop button sensor is triggered

Does nothing but print a message
*/
func onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
}

/*
	Activates when the timer ends
	Either the door should close or the elevator is stuck

Input: the elevator object

Returns: the elevator object with updated state
*/
func onTimerEnd(elevator Elevator) Elevator {

	switch elevator.Behaviour {
	case EB_DoorOpen:
		if !elevator.Obstruction {
			elevator.Stuck = false
			direction, behaviour := chooseElevatorDirection(elevator)
			elevator.Direction = direction
			elevator.Behaviour = behaviour
			if elevator.Behaviour == EB_DoorOpen {
				elevator = clearCallsAtCurrentFloor(elevator)
			}
		} else {
			elevator.Stuck = true
		}
	case EB_Moving:
		elevator.Stuck = true
	}
	return elevator
}
