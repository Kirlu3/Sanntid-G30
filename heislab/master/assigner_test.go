package master

import (
	"log"
	"os"
	"testing"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func TestAssigner(t *testing.T) {
	elevators := [config.N_ELEVATORS]slave.Elevator{
		{
			Floor:     1,
			Direction: -1,
			Calls: [4][3]bool{
				{false, false, false},
				{false, false, false},
				{false, false, false},
				{false, false, false},
			},
			Behaviour: 2,
			Stuck:     false,
			ID:        0,
		},
		{
			Floor:     0,
			Direction: 0,
			Calls: [4][3]bool{
				{false, false, false},
				{false, false, false},
				{false, false, false},
				{false, false, false},
			},
			Behaviour: 0,
			Stuck:     false,
			ID:        1,
		},
		{
			Floor:     0,
			Direction: 0,
			Calls: [4][3]bool{
				{false, false, false},
				{false, false, false},
				{false, false, false},
				{false, false, false},
			},
			Behaviour: 0,
			Stuck:     false,
			ID:        2,
		},
	}

	callsToAssign := AssignCalls{
		Calls: Calls{
			HallCalls: [4][2]bool{
				{true, false},
				{true, false},
				{true, false},
				{false, false},
			},
			CabCalls: [3][4]bool{
				{false, false, false, true},
				{false, false, true, true},
				{false, true, false, false},
			},
		},
		AliveElevators: [3]bool{false, true, false},
	}

	if err := os.Chdir("../../"); err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}
	assign(elevators, callsToAssign.Calls, callsToAssign.AliveElevators)

}
