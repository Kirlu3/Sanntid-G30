changed init in backupAckRx to: consider if this makes sense
wantReassignment = true

TODO heter de det samme inne i slave funksjonen?? 

consider renaming peers to alive

bugs: 
elevator takes 3 seconds to clear calls in order it is traveling even if no such calls exist  
elevator doesnt clear hall calls at current floor after becoming unobstructed, maybe related to floorarrival method of clearing calls?  
