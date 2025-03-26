package slave

import (
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

func Slave(
	id string,
	offlineCallsToSlaveChan <-chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	offlineSlaveBtnToMasterChan chan<- ButtonMessage,
	offlineSlaveStateToMasterChan chan<- Elevator,
	startSendingBtnOfflineChan <-chan struct{},
	startSendingStateOfflineChan <-chan struct{},
) {

	ID, _ := strconv.Atoi(id)

	drvBtnChan := make(chan elevio.ButtonEvent)
	drvNewFloorChan := make(chan int)
	drvObstrChan := make(chan bool)
	drvStopChan := make(chan bool)

	slaveStateToMasterChan := make(chan Elevator, 2)
	callsFromMasterChan := make(chan [config.N_FLOORS][config.N_BUTTONS]bool)

	var timer *time.Timer = time.NewTimer(0)
	<-timer.C

	go elevio.PollButtons(drvBtnChan)
	go elevio.PollFloorSensor(drvNewFloorChan)
	go elevio.PollObstructionSwitch(drvObstrChan)
	go elevio.PollStopButton(drvStopChan)

	go buttonPressTx(drvBtnChan, offlineSlaveBtnToMasterChan, startSendingBtnOfflineChan, ID)
	go slaveStateTx(slaveStateToMasterChan, offlineSlaveStateToMasterChan, startSendingStateOfflineChan)
	go callsFromMasterRx(callsFromMasterChan, offlineCallsToSlaveChan, ID)

	go fsm(ID, slaveStateToMasterChan, callsFromMasterChan, drvNewFloorChan, drvObstrChan, drvStopChan, timer)
}
