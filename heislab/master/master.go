package master

import (
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/alive"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

/*
# The main function of the master module. Initializes all channels and goroutines.
*/
func Main(
	initialCalls Calls,
	Id int,
	offlineCallsToSlaveChan chan<- [config.NumElevators][config.NumFloors][config.NumBtns]bool,
	offlineSlaveBtnToMasterChan <-chan slave.ButtonMessage,
	offlineSlaveStateToMasterChan <-chan slave.Elevator,
) {
	callsUpdateChan := make(chan struct {
		Calls   Calls
		AddCall bool
	}, 2)

	callsToAssignChan := make(chan struct {
		Calls          Calls
		AliveElevators [config.NumElevators]bool
	})

	slaveStateUpdateChan := make(chan slave.Elevator)
	callsToSlaveChan := make(chan [config.NumElevators][config.NumFloors][config.NumBtns]bool)
	callsToBackupsTxChan := make(chan Calls)
	enableMasterTxChan := make(chan bool)

	go alive.Transmitter(config.MasterUpdatePort, strconv.Itoa(Id), enableMasterTxChan)

	go callsFromBackupsRx(callsUpdateChan, callsToAssignChan, callsToBackupsTxChan, initialCalls, Id)
	go callsToBackupsTx(callsToBackupsTxChan, initialCalls, Id)

	go assigner(slaveStateUpdateChan, callsToAssignChan, callsToSlaveChan, Id)

	go buttonPressRx(callsUpdateChan, offlineSlaveBtnToMasterChan)
	go slaveStateUpdateRx(slaveStateUpdateChan, callsUpdateChan, offlineSlaveStateToMasterChan)
	go callsToSlavesTx(callsToSlaveChan, offlineCallsToSlaveChan)

}
