package Backup

import "time"

// detect master disconnections and transition to master when needed

const (
	DEADLINE_MASTER = time.Second*3
)



func timer(timeSinceMessage chan time.Duration) {




	for{
		time := <- timeSinceMessage
		
		if time >= DEADLINE_MASTER {

				//-- TODO --
				// 1. check priority and how many other elevators that are alive
				// 2. if priority is highest or noone else is alive on network --> transition to master 
	
		}
	}
	
}