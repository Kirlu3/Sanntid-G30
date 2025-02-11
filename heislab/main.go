package main

import (
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func main() {
	go Slave.Slave()
	select {}
}
