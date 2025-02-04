package main

import (
	Backup "github.com/Kirlu3/Sanntid-G30/heislab/backup"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	Master "github.com/Kirlu3/Sanntid-G30/heislab/master"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func main() {
	N_FLOORS := 4
	elevio.Init("localhost:15657", N_FLOORS)
	go Slave.Slave()
	go Master.Master()
	go Backup.Backup()

	select {}
}
