some concerns:

sending slices between go routines sends a reference to the slice, which kind of defeats the purpose of message passing?
solution could be to explictly copy, though this is somewhat complicated for complex structs, like a struct which has a slice of structs, which themselves have slice members
Otherwise i think we have to be very careful
it might be worth making functions to deep copy most common structs, e.g. the full worldview
go has built in copy() function which performs shallow copy, deep copy seems more complicated
for now i have imported "github.com/mohae/deepcopy" which should be able to deepcopy pretty much anything i think?

i think it is fine to assign orders without waiting for the lights, doing so seems to make the code much simpler, as you can always just assign the full state without worrying about what is confirmed and not

currently my plan is to store the full master worldview in the stateManager go routine, this means that all reads/writes goes through this module, so it need a lot of channels

when master phase ends a lot of go routines should end and channels should be shut down, how to do this?
just shut down all channels and make go routines return/panic, very uncontrolled but might be fine since we are not really looking to store any information since we are just going back to backup phase
alternative is more controlled shutdown, but i think this requires much more advanced logic, i.e. channels dedicated to shutting down go routines

i have taken the liberty of adding an Id field to the Slave.Elevator
i think it would be nice for indexing if every id is a string in the interval "0" to "9"
this way the worldview struct can have an array of 10 elevators and when we get message from elevator "i" we can update Elevators[i]
is this ok? objections?

in general if you have any objections to the data format and the way stuff is passed around it should be adressed asap, because it would be very annoying to change later

some places maybe buffered channels would be good to prevent blocking


Want reassignment on: callUpdate, backupUpdate, anything else??


## TODO:
fix messages between slaveComm and assigner, backupAckRx

fix graceful shutdown of master mode

fix init of fsm (set ligths to 0 maybe?)

how do the slaves handle losing their master?

apparently go passes structs and arrays by value so deepcopy is probably not necessary?