package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/backup"
	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func main() {
	id := flag.String("id", "inv", "id of this elevator")
	serverPort := flag.String("port", "15657", "port to communicate with elevator")
	flag.Parse()

	if *id == "inv" {
		panic("please specify an id in [0, NumElevators] with -id=x")
	}

	idInt, err := strconv.Atoi(*id)

	if err != nil {
		panic("id must be an integer")
	}

	if idInt < 0 || idInt >= config.NumElevators {
		panic("id must be in [0, NumElevators]")
	}

	serverAddress := fmt.Sprintf("localhost:%s", *serverPort)
	elevio.Init(serverAddress, config.NumFloors)

	offlineCallsToSlaveChan := make(chan [config.NumElevators][config.NumFloors][config.NumBtns]bool)
	offlineSlaveBtnToMasterChan := make(chan slave.ButtonMessage)
	offlineSlaveStateToMasterChan := make(chan slave.Elevator)

	startSendingBtnOfflineChan := make(chan struct{})
	startSendingStateOfflineChan := make(chan struct{})

	slave.Main(idInt, offlineCallsToSlaveChan, offlineSlaveBtnToMasterChan, offlineSlaveStateToMasterChan, startSendingBtnOfflineChan, startSendingStateOfflineChan)

	backedUpCalls := backup.Run(idInt)

	startSendingBtnOfflineChan <- struct{}{}
	startSendingStateOfflineChan <- struct{}{}

	master.Main(backedUpCalls, idInt, offlineCallsToSlaveChan, offlineSlaveBtnToMasterChan, offlineSlaveStateToMasterChan)

	select {}
}
