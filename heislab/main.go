package main

import (
	"./driver-go/elevio"

	Slave "Slave"
)

func main() {
	N_FLOORS := 4
	elevio.Init("localhost:15657", N_FLOORS)
	go Slave.Slave()

	select {}
}
