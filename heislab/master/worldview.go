package master

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type Calls struct {
	HallCalls [config.N_FLOORS][config.N_BUTTONS - 1]bool
	CabCalls  [config.N_ELEVATORS][config.N_FLOORS]bool 
}


// ok this is kind of stupid but i dont know what to do about it

type BackupCalls struct {
	Calls Calls
	Id    int
}

type AssignCalls struct {
	Calls          Calls
	AliveElevators [config.N_ELEVATORS]bool
}

type UpdateCalls struct {
	Calls   Calls
	AddCall bool
}
