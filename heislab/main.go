package main

import (
	"./driver-go/elevio"
	"./slave"
)

func main() {
	N_FLOORS := 4
	elevio.Init("localhost:15657", N_FLOORS)
	go slave.Slave()

	select {}
}
