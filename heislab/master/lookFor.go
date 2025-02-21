package master

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

// consider using the masterWorldViewRx instead
func lookForOtherMasters(endMasterPhase chan<- struct{}, masterWorldViewRx <-chan slave.WorldView, ownId string, mergeState chan<- slave.WorldView) {
	// lookForMastersLoop:
	for {
		select {
		case masterWorldView := <-masterWorldViewRx:
			if masterWorldView.OwnId != ownId {
			mergeState <- masterWorldView
			}
		}

	}
}
