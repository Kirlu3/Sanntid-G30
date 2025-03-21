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
	drv_NewFloorChan <-chan int,
	drv_ObstrChan <-chan bool,
	drv_StopChan <-chan bool,
	timerDurationChan chan int,
	timer *time.Timer,
) {

	//initialize elevator
	var elevator Elevator
	elevator.ID = ID
	updateLights(elevator.Calls)

	newElevator := initElevator(elevator)
	elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

	for {
		fmt.Println("FSM:New Loop")
		activateElevatorIO(elevator)
		select {
		case newCalls := <-callsFromMasterChan:
			fmt.Println("Slave: Updating orders")
			elevator.Calls = newCalls
			newElevator = fsm_onNewCalls(elevator)
			elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

		case floor := <-drv_NewFloorChan:
			fmt.Println("FSM: Floor arrival", floor)
			newElevator = fsm_onFloorArrival(floor, elevator)
			elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

		case obstr := <-drv_ObstrChan:
			newElevator = fsm_onObstruction(obstr, elevator)
			elevator = updateElevatorState(newElevator, elevator, slaveStateToMasterChan, timerDurationChan)

		case <-drv_StopChan:
			fsm_onStopButtonPress()

		case <-timer.C:
			fmt.Println("FSM: Timer end")
			newElevator = fsm_onTimerEnd(elevator)
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
	fmt.Println("onInit")
	elevator.Direction = D_Down
	elevator.Behaviour = EB_Moving
	fmt.Println("offInit")
	return elevator
}

/*
	Activates when the elevator receives new calls

Input: the old elevator object with updated calls

Returns: the new elevator object with updated direction and behaviour
*/
func fsm_onNewCalls(elevator Elevator) Elevator {
	fmt.Println("onRequest, behaviour:", elevator.Behaviour)
	switch elevator.Behaviour {
	case EB_Idle:
		direction, behaviour := chooseElevatorDirection(elevator)
		elevator.Direction = direction
		elevator.Behaviour = behaviour
		if elevator.Behaviour == EB_DoorOpen {
			elevator = clearCallsAtCurrentFloor(elevator)
		}
	case EB_DoorOpen:
		elevator = clearCallsAtCurrentFloor(elevator)
	}

	return elevator
}

/*
	Activates when the elevator floor sensor is triggered

Input: the elevator and the new floor

Returns: the elevator with the floor updated
*/
func fsm_onFloorArrival(newFloor int, elevator Elevator) Elevator {
	if !elevator.Obstruction {
		elevator.Stuck = false //if the elevator arrives at a floor, it is not stuck
	}
	fmt.Println("onFloorArrival")
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
func fsm_onObstruction(obstruction bool, elevator Elevator) Elevator {
	fmt.Println("onObstruction")
	elevator.Obstruction = obstruction
	return elevator
}

/*
	Activates when the stop button sensor is triggered

Does nothing but print a message
*/
func fsm_onStopButtonPress() {
	fmt.Println("You pressed the stop button :)")
}

/*
	Activates when the timer ends
	Either the door should close or the elevator is stuck

Input: the elevator object

Returns: the elevator object with updated state
*/
func fsm_onTimerEnd(elevator Elevator) Elevator {

	switch elevator.Behaviour {
	case EB_DoorOpen:
		fmt.Println("FSM:onTimerEnd DO")

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
		fmt.Println("FSM:onTimerEnd M")
		elevator.Stuck = true
	}
	return elevator
}
