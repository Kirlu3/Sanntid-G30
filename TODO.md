changed init in backupAckRx to: consider if this makes sense
wantReassignment = true

TODO heter de det samme inne i slave funksjonen?? 

consider renaming peers to alive

consider reassigning every time a state update happens (why not?)

consider changing either everything, or everything in fsm to include or not include prefixes

What is the purpose of assign all calls to master? For the case of no alive elevators? consider just setting master elevator to alive and then assigning in the normal way


Master:
The master calls recursive go routines, we should try to avoid this. This is especially true for the backupcoordinator which I think has some bloat still. Some of the routines don't need to exist, the others I think should be called from master.go. I want to rework this function a bit, left some comments.

change shouldReassign bool. It's unnecessary in this case I think, as we might as well reassign every time we get an updated state or calls. 

bugs: 
elevator doesnt clear hall calls at current floor after becoming unobstructed, maybe related to floorarrival method of clearing calls?  

TODO: change obstruction to not stop and open door