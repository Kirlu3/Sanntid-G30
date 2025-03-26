package master

import (
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/alive"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

func Master(
	initialCalls Calls,
	id_string string,
	offlineCallsToSlaveChan chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	offlineSlaveBtnToMasterChan <-chan slave.ButtonMessage,
	offlineSlaveStateToMasterChan <-chan slave.Elevator,
) {
	Id, _ := strconv.Atoi(id_string)

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
