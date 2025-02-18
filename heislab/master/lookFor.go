package master

import (
	"slices"

	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
)

// consider using the masterWorldViewRx instead
func lookForOtherMasters(endMasterPhase chan struct{}, masterUpdateCh chan peers.PeerUpdate, ownId string) {
	// lookForMastersLoop:
	for {
		select {
		case masterUpdate := <-masterUpdateCh:
			if slices.Min(masterUpdate.Peers) < ownId {
				// end the master phase, but the other master needs to get our state first!
			}
			if slices.Max(masterUpdate.Peers) > ownId {
				// make sure the stateManager gets the mergeState message
			}
		}
	}
}
