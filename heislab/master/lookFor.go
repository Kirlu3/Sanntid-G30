package master

// If we detect another master, forward this information to ???
func lookForOtherMasters(otherMasterCallsCh chan<- BackupCalls, ownId int, masterCallsRx <-chan BackupCalls) {
	for otherMasterCalls := range masterCallsRx {
		if otherMasterCalls.Id != ownId {
			otherMasterCallsCh <- otherMasterCalls
		}
	}
}
