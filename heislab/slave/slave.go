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
) {

	ID, _ := strconv.Atoi(id)

	//initialize channels
	drv_BtnChan := make(chan elevio.ButtonEvent)
	drv_NewFloorChan := make(chan int)
	drv_ObstrChan := make(chan bool)
	drv_StopChan := make(chan bool)

	slaveStateToMasterChan := make(chan Elevator, 2)
	callsFromMasterChan := make(chan [config.N_FLOORS][config.N_BUTTONS]bool)

	timerDurationChan := make(chan int)

	//initialize timer
	var timer *time.Timer = time.NewTimer(0)
	<-timer.C
	go resetTimer(timerDurationChan, timer)

	//initialize sensors
	go elevio.PollButtons(drv_BtnChan)
	go elevio.PollFloorSensor(drv_NewFloorChan)
	go elevio.PollObstructionSwitch(drv_ObstrChan)
	go elevio.PollStopButton(drv_StopChan)

	//initialize network
	go buttonPressTx(drv_BtnChan, offlineSlaveBtnToMasterChan, ID)
	go slaveStateTx(slaveStateToMasterChan, offlineSlaveStateToMasterChan)
	go callsFromMasterRx(callsFromMasterChan, offlineCallsToSlaveChan, ID)

	//initialize fsm
	go fsm(ID, slaveStateToMasterChan, callsFromMasterChan, drv_NewFloorChan, drv_ObstrChan, drv_StopChan, timerDurationChan, timer)
}
