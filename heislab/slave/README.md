Slave module
===================

Consists of the following files:
- elevator.go
- fsm.go
- io.go
- masterComm.go
- slave.go
- timer.go

## elevator.go
Contains the elevator struct which tracks the state of the elevator. 
Also contains a function to check the validity of an elevators state and a function to update the current elevator object along with interfacing with IO and sending elevator state change messages.
Also contains support functions for the FSM that deal with manipulating and getting information from the elevator object.

## fsm.go
Contains the routine that runs the FSM in addition to functions to handle each incoming event. The functions are called when the elevator detects the corresponding event for each fsm function. These are as follows:
- Initialization
- Updated requests
- The elevator has arrived at a floor
- There has been a change in the door obstruction switch
- The stop button has been pressed
- The timer has timed out

## io.go
Contains a function to activate the IO of the elevator based on its current state. This is called at the start of each loop of the FSM.
Also contains a function to turn on or off the order lights based on an incoming lights update from the master. 

## masterComm.go
Contains three routines. Two senders and one receiver.
The first sending routine sends button presses to the master and waits for acknowledgement before trying again.
The second sending routine broadcasts the most recent elevator state to the master.
The receiving routine listens to order assignment UDP broadcasts from the master and translates them to updated order assignments and lights.

## slave.go
Contains a single function that initializes all channels and go routines between other parts of the slave module. This is essentially the main file for the slave module.