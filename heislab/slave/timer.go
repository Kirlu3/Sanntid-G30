package slave

import "time"

var doorOpenDuration = time.Second * 3

func timer(t_start chan bool, t_end *time.Timer) {
	for {
		if <-t_start {
			t_end.Reset(doorOpenDuration)
		}
	}
}
