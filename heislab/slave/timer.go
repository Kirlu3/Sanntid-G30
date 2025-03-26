package slave

import "time"

/*
# Go routine that resets the timer when it receives a new duration

Input: timerDurationChan, timer

timerDurationChan: receives the duration of the timer

timer: the timer to reset
*/
func resetTimer(timerDurationChan chan int, timer *time.Timer) {
	for seconds := range timerDurationChan {
		timer.Reset(time.Second * time.Duration(seconds))
	}
}
