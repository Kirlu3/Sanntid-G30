package main

import (
	//"github.com/Kirlu3/Sanntid-G30/heislab/backup"
	"fmt"

	//"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func main() {
	id := "2"
	//N_FLOORS := 4
	//elevio.Init("localhost:15657", N_FLOORS)
	//go slave.Slave()
	//go backup.Backup(id)

	state := slave.WorldView{
		Elevators: [10]slave.Elevator{
			{Floor: 2, Direction: slave.D_Up, Requests: [4][3]bool{}, Behaviour: slave.EB_Moving, Stuck: false, Id: "0"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "1"},
			{Floor: 3, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "2"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "3"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "4"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "5"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "6"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "7"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "8"},
			{Floor: 0, Direction: 0, Requests: [4][3]bool{}, Behaviour: 0, Stuck: false, Id: "9"},
		},
		OwnId:          id,
		HallCalls: [4][2]bool{
			{false, false},
			{true, false},
			{false, false},
			{false, false},
			},
		CabCalls: [10][4]bool{
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
			{false, false, false, false},
		},
		AliveElevators: [10]bool{true, true, false, false, false, false, false, false, false, false},
	}

	assignment := master.Assign(state)
	

	fmt.Println(assignment)


	//select {}
}