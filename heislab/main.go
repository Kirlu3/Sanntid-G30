package main

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/backup"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func main() {
	id := "2"
	N_FLOORS := 4
	elevio.Init("localhost:15657", N_FLOORS)
	go Slave.Slave()
	go backup.Backup(id)

	select {}
}
