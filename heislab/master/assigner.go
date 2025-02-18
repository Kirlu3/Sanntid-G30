package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func assignOrders(stateToAssign chan slave.WorldView, orderAssignments chan [][]int) {
	for {
		select {
		case state := <-stateToAssign:
			fmt.Printf("state: %v\n", state)
			var assignments [][]int // TODO: call the D cost algo
			orderAssignments <- assignments
		}
	}
}
