package main

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func main() {
	N_FLOORS := 4
	elevio.Init("localhost:15657", N_FLOORS)
	go Slave.Slave()

	select {}
}
