package slave

import (
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

// Initialization and main loop of the slave module
func Slave(
	id string,
	offlineCallsToSlaveChan <-chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	offlineSlaveBtnToMasterChan chan<- ButtonMessage,
	offlineSlaveStateToMasterChan chan<- Elevator,
	startSendingBtnOfflineChan <-chan struct{},
	startSendingStateOfflineChan <-chan struct{},
) {

	ID, _ := strconv.Atoi(id)

	//initialize channels
	drvBtnChan := make(chan elevio.ButtonEvent)
	drvNewFloorChan := make(chan int)
	drvObstrChan := make(chan bool)
	drvStopChan := make(chan bool)

	slaveStateToMasterChan := make(chan Elevator, 2)
	callsFromMasterChan := make(chan [config.N_FLOORS][config.N_BUTTONS]bool)

	timerDurationChan := make(chan int)

	//initialize timer
	var timer *time.Timer = time.NewTimer(0)
	<-timer.C
	go resetTimer(timerDurationChan, timer)

	//initialize sensors
	go elevio.PollButtons(drvBtnChan)
	go elevio.PollFloorSensor(drvNewFloorChan)
	go elevio.PollObstructionSwitch(drvObstrChan)
	go elevio.PollStopButton(drvStopChan)

	//initialize network
	go buttonPressTx(drvBtnChan, offlineSlaveBtnToMasterChan, startSendingBtnOfflineChan, ID)
	go slaveStateTx(slaveStateToMasterChan, offlineSlaveStateToMasterChan, startSendingStateOfflineChan)
	go callsFromMasterRx(callsFromMasterChan, offlineCallsToSlaveChan, ID)

	//initialize fsm
	go fsm(ID, slaveStateToMasterChan, callsFromMasterChan, drvNewFloorChan, drvObstrChan, drvStopChan, timerDurationChan, timer)
}
