package master

import (
	"fmt"
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func Master(
	initCalls Calls,
	Id int,
	offlineCallsToSlaveChan chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	offlineSlaveBtnToMasterChan <-chan slave.ButtonMessage,
	offlineSlaveStateToMasterChan <-chan slave.Elevator,
) {
	fmt.Println(Id, "entered master mode")

	callsUpdateChan := make(chan struct {
		Calls   Calls
		AddCall bool
	}, 2)

	callsToAssignChan := make(chan struct {
		Calls          Calls
		AliveElevators [config.N_ELEVATORS]bool
	})

	slaveStateUpdateChan := make(chan slave.Elevator)
	callsToSlaveChan := make(chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool)
	enableMasterTxChan := make(chan bool)

	go peers.Transmitter(config.MasterUpdatePort, strconv.Itoa(Id), enableMasterTxChan)

	go backupCoordinator(callsUpdateChan, callsToAssignChan, initCalls, Id)
	go assignCalls(slaveStateUpdateChan, callsToAssignChan, callsToSlaveChan)

	go buttonPressRx(callsUpdateChan, offlineSlaveBtnToMasterChan)
	go slaveStateUpdateRx(slaveStateUpdateChan, callsUpdateChan, offlineSlaveStateToMasterChan)
	go callsToSlavesTx(callsToSlaveChan, offlineCallsToSlaveChan)

	// the program is crashed and restarted when it should go back to backup mode
	select {}
}
