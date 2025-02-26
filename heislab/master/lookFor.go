package master

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

// If we detect another master, forward this information to ???
func lookForOtherMasters(masterCallsRx <-chan slave.BackupCalls, ownId int, mergeState chan<- slave.BackupCalls) {
	for {
		select {
		case otherMasterCalls := <-masterCallsRx:
			if otherMasterCalls.Id != ownId {
				mergeState <- otherMasterCalls
			}
		}
	}
}
