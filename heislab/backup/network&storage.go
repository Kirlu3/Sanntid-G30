package Backup

import "time"

// Network & storage - one to store/resend the master’s worldview

func networkStorage(timeSinceMessage chan time.Duration){
	timeBefore := time.Now()

	// receive master's worldview

	timeAfter := time.Now()

	timeSinceMessage <- (timeAfter - timeBefore)

	
	
} 