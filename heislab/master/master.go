package Master

// the master can make BIG decisions (like turning on lights) 
// if it has received conformation that all other alive backups mirror its own state (but does this make more sense with tcp?)
// if no message from x for some time: mark x as dead
// reassign orders assigned to x (SMALL decision, can be done anytime?)


func Master() {
	// EXAMPLE OF POSSIBLE THREADS IN MASTER
	// go sendStateToBackups()
	// go receiveStateFromBackups()
	// go sendMessagesToSlaves() // orders + anything else?
	// go receiveMessagesFromSlaves() // all relevant info from slave
	// maybe we store some list of lights we want to turn on/other big decisions we want to make
	// when we get the green light from receiveStateFromBackups we execute?

	// go lookForOtherMasters() // and handle potential master merging? 
	// could be spawned from receive from backups 
	// (or spawned here and receive message from receiveStateFromBackups if it needs to do something?)
	// go assignOrders
	// go establishConnectionsToSlaves





	// end all goroutines and return (to backup state) (if we learn that there are other masters with higher priority?)
	// does this master/backups structure make sense?

}
