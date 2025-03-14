package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

// If we detect another master, forward this information to ???
func lookForOtherMasters(otherMasterCallsCh chan<- struct {Calls Calls; Id int}, ownId int) {
	masterCallsRx := make(chan struct{Calls Calls; Id int})
	go bcast.Receiver(config.MasterCallsPort, masterCallsRx)
	for otherMasterCalls := range masterCallsRx {
		if otherMasterCalls.Id != ownId {
			fmt.Println("found other master")
			otherMasterCallsCh <- otherMasterCalls
		}
	}
}
