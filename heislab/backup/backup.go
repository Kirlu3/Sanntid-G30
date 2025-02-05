package Backup


// maybe make a config file where we can change all parameters
// set the time until it is time to become the master
// i guess each node should have an id and associated address/PORT e.g. id = x, port = 6x009?

// find a way to init the program with an id

// we should be operating with the 

var _BACKUP_TAKEOVER_SECONDS int = 5


func Backup() {
	for {
		var NODE_ID int = 2
		var NODE_PORT int = ID_to_PORT(NODE_ID)
		// what is our local state on startup?
		// init to DONT_KNOW

		// who is the master?
	
		// start in backup phase
		// wait to hear from the master
		// if we dont hear from master for _BACKUP_TAKEOVER_SECONDS seconds we break this loop
		for {

			master_state_bytes := listen_from_master()
			master_state_struct := decode_master_state(master_state_bytes)
			// the message contains some information which tells us if this is a newer version or not
			// do we need one go routine for sending and one for receiving? or alternate?
			// or is this all possible messages from any master/backup?
			// acceptance test of the master state? make sure the message actually makes sense (but what do we do if it doesnt)
	
			// should we do something if we received a message from a different master?
			// we definetively need to store who is the master/what address we are sending to
			// or no we just broadcast at our address, the master sends to all addresses
			// but it is probably cumbersome for the master to read from all addresses
	
			// WHEN THE MASTER DIES:
			// the master has sent us id of all alive elevators, lowest should become new master
			// if we are lowest become
	
	
	
			break
		}
	
	
		// wait for master message phase
		if (anyLessThan(other_known_ids, our_id)) {
			// listen for master message among ALL POSSIBLE other_ids
		}
	
	
		// become master phase
		// become master and block the backup code
		Master.Master(full state i guess)
	}

}
