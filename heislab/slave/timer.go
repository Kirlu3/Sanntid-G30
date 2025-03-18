package slave

import "time"

/*
	resetTimer resets the timer with a duration in seconds given by timerDurationChan.
*/
func resetTimer(timerDurationChan chan int, timer *time.Timer) {
	for seconds := range timerDurationChan {
		timer.Reset(time.Second * time.Duration(seconds))
	}
}
