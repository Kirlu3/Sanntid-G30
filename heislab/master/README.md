Master module
===================

Consists of the following files:
- assigner.go
- backupComm.go
- master.go
- slaveComm.go

## assigner.go
Contains a routine and functions to assign orders upon receiving new confirmed orders or new elevator states.
It assigns orders using the given hall_request_assigner before sending them to be broadcast to the slaves.

## backupComm.go
Contains two go routines. One broadcasts the most recent calls to be updated to the backups.
The other listens to the backups for acknowledgment on the most recent calls. This routine also listens for calls from other masters to ensure all calls are backed up in case of multiple masters before one is allowed to crash. 
Also contains pure functions used by these routines. 

## master.go
Contains a function that initializes all necessary channels and starts all necessary go routines. This is essentially the main file for the master module.

## slaveComm.go
Contains three routines to deal with communication with the slaves. One that listens to the state broadcasts of the slaves and sends updates to the assigner. One that listens to button presses and sends them to be backed up. Finally one to send assigned calls to the slaves.