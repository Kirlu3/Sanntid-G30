# Sanntid-G30
Sanntidsprosjekt gruppe 30 V25

TODO:
-Fix communication
-Add channel directions to *every* function
-Add comments to *every* function
    -Stuff like output and input and purpose

-Watchdog with program that restarts on a crash (would fix the occasional crashes on startup)

-Change config to include everything necessary, including: Base IP

-fix messages between slaveRx and assignOrders, backupAckRx

-fix graceful shutdown of master mode

-how do the slaves handle losing their master?
    -they don't notice, keep sending and receiving from the same place

-apparently go passes structs and arrays by value so deepcopy is probably not necessary?

-FIX: why do lights cab lights turn off on obstruction? something related to obstruction seems to break the elevator

-FIX: if an elevator has a single call in the direction it is coming from, it failt to clear it

-Consider: with a lot of packet loss so messages take a long time and new assignments are sent before a previous clear was received

-FIX: maybe bug in slave/comm.go line 60 

-FIX: sometimes when fucking around too much the system gets fucked beyond saving: hall calls are assigned to elevator 0, even though it does not exist? refuses cab calls:
As: state: {[{0 0 [[false false false] [false false false] [false false false] [false false false]] 0 false 0} {0 0 [[false false false] [false false false] [false false false] [false false false]] 0 false 0} {2 -1 [[false false false] [false false false] [false false false] [false false false]] 1 false 2}]  [[false false] [true true] [true true] [false true]] [[false false false false] [true true true false] [false false false false]] [false true false]}
ST: New orders sent
[[[false false false] [true true false] [true true false] [false true false]] [[false false false] [false false false] [false false false] [false false false]] [[false false false] [false false false] [false false false] [false false false]]]
-Consider: making pretty print for the the above for easier debugging

The problem in the above assignment appears to be that somehow id of elevator 1 has been set to 0, when is Elevator.Id field updated? Is the id wrong for the slave or just the master/assigner
can we make the assignments without using the id field? if we do will the id being wrong have other consequences?
-update: i think it is just an init problem in the assigner: elev1 isnt moving so doesnt get any state updates so doesnt get to set its id (init is 0) to 1. Therefore it cant get assignments and will never move
suggested fix: set id fields on init. 
Also consider: not using the id fields to do assingments

Ensure: That all cab orders from an offline elevator is sent to the corresponding master before it is potentially crashed by encountering another master with higher priority
