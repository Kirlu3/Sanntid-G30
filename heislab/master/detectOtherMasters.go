package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

/*
The routine checks for the presence of other masters in the system. If another master is detected, the calls and id from that master are forwarded through otherMasterUpdateChan.
*/
func detectOtherMasters(otherMasterUpdateChan chan<- struct {Calls Calls; Id int}, ownId int) {
	masterCallsRx := make(chan struct{Calls Calls; Id int})
	go bcast.Receiver(config.MasterCallsPort, masterCallsRx)
	for otherMasterCalls := range masterCallsRx {
		if otherMasterCalls.Id != ownId {
			fmt.Println("found other master")
			otherMasterUpdateChan <- otherMasterCalls
		}
	}
}
