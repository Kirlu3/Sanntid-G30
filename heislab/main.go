package main

import (
	"flag"
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/backup"
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

// the program should be called with go run heislab/main.go -id=x -port=5590x
func main() {
	// id := os.Args[1:][0]
	id := flag.String("id", "inv", "id of this elevator")
	serverPort := flag.String("port", "15657", "port to communicate with elevator")
	flag.Parse()

	if *id == "inv" {
		panic("please specify an id in [0, N_Elevators) with -id=x")
	}

	serverAddress := fmt.Sprintf("localhost:%s", *serverPort)
	elevio.Init(serverAddress, config.N_FLOORS)

	// Channels for offline communication
	offlineCallsToSlaveChan := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	offlineSlaveBtnToMasterChan := make(chan slave.ButtonMessage)
	offlineSlaveStateToMasterChan := make(chan slave.Elevator)

	slave.Slave(*id, offlineCallsToSlaveChan, offlineSlaveBtnToMasterChan, offlineSlaveStateToMasterChan)

	backedUpCalls := backup.Backup(*id)
	master.Master(backedUpCalls, *id, offlineCallsToSlaveChan, offlineSlaveBtnToMasterChan, offlineSlaveStateToMasterChan)

	select {}
}
