changed init in backupAckRx to: consider if this makes sense
wantReassignment = true
Master:
The master calls recursive go routines, we should try to avoid this. This is especially true for the backupcoordinator which I think has some bloat still. Some of the routines don't need to exist, the others I think should be called from master.go. I want to rework this function a bit, left some comments.

TODO: changing either everything, or everything in fsm to include or not include prefixes

TODO: realized the door timer should be set differently. Per now it needs 3 seconds without anything happening in the fsm to close. 

TODO: ensure we all agree with naming and ensure we're consistent and it makes sense everywhere.

TODO: make own FAT test based on project specs

TODO: general cleanup and code quality when we have time (thing I realized, fix all the print statements)

bugs: 
elevator doesnt clear hall calls at current floor after becoming unobstructed, maybe related to floorarrival method of clearing calls? <- test this

TODO: consider if the assigner should wait for first calls message before trying to assign (if not it would only assign nothing)
