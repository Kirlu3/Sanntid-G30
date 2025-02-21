package master

import (
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

// consider using the masterWorldViewRx instead
func lookForOtherMasters(endMasterPhase chan struct{}, masterWorldViewRx chan slave.WorldView, ownId string) {
	// lookForMastersLoop:
	for {
		select {
		case masterWorldView := <-masterWorldViewRx:
			if masterWorldView.OwnId < ownId {
				// end the master phase, but the other master needs to get our state first!
				// end if the other master sends a state that we find acceptable
			} else if masterWorldView.OwnId > ownId {
				// make sure the stateManager gets the mergeState message
				// make sure out state is acceptable so the other master can end
			} // if ids are equal we received our own message
		}
	}
}
