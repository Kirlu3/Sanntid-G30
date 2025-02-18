package slave

import "time"

func timer(t_start chan int, t_end *time.Timer) {
	for {
		a := <-t_start
		t_end.Reset(time.Second * time.Duration(a))
	}
}
