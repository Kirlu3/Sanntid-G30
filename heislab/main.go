package main

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/backup"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func main() {
	N_FLOORS := 4
	elevio.Init("localhost:15657", N_FLOORS)
	go Slave.Slave()
	go master.Master()
	go backup.Backup()

	select {}
}
