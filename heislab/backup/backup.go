package Backup

import "time"

// --- MODULES --- //

// Timer - detect master disconnections and transition to master when needed
// Network & storage - one to store/resend the master’s worldview


func Backup() {
	timeSinceMessage := make(chan time.Duration)

	go timer(timeSinceMessage)
	go networkStorage(timeSinceMessage)

}
