package slave

import (
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

/*
# The main function of the slave module. Initializes all channels and goroutines.
*/
func Main(
	Id int,
	offlineCallsToSlaveChan <-chan [config.NumElevators][config.NumFloors][config.NumBtns]bool,
	offlineSlaveBtnToMasterChan chan<- ButtonMessage,
	offlineSlaveStateToMasterChan chan<- Elevator,
	startSendingBtnOfflineChan <-chan struct{},
	startSendingStateOfflineChan <-chan struct{},
) {
	drvBtnChan := make(chan elevio.ButtonEvent)
	drvNewFloorChan := make(chan int)
	drvObstrChan := make(chan bool)
	drvStopChan := make(chan bool)

	slaveStateToMasterChan := make(chan Elevator, 2)
	callsFromMasterChan := make(chan [config.NumFloors][config.NumBtns]bool)

	var elevatorTimer *time.Timer = time.NewTimer(0)
	<-elevatorTimer.C

	go elevio.PollButtons(drvBtnChan)
	go elevio.PollFloorSensor(drvNewFloorChan)
	go elevio.PollObstructionSwitch(drvObstrChan)
	go elevio.PollStopButton(drvStopChan)

	go buttonPressTx(drvBtnChan, offlineSlaveBtnToMasterChan, startSendingBtnOfflineChan, Id)
	go slaveStateTx(slaveStateToMasterChan, offlineSlaveStateToMasterChan, startSendingStateOfflineChan)
	go callsFromMasterRx(callsFromMasterChan, offlineCallsToSlaveChan, Id)

	go fsm(Id, slaveStateToMasterChan, callsFromMasterChan, drvNewFloorChan, drvObstrChan, drvStopChan, elevatorTimer)
}
