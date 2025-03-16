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
	masterToSlaveCalls_offlineChann <-chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	slaveToMasterBtn_offlineCha chan<- ButtonMessage,
	slaveToMasterElevState_offlineChan chan<- Elevator,
) {

	ID, _ := strconv.Atoi(id)

	//initialize channels
	drv_BtnChan := make(chan elevio.ButtonEvent)
	drv_NewFloorChan := make(chan int)
	drv_ObstrChan := make(chan bool)
	drv_StopChan := make(chan bool)
	startTimerChan := make(chan int)

	elevatorUpdateChan := make(chan Elevator, 2)
	callsReceiverChan := make(chan [config.N_FLOORS][config.N_BUTTONS]bool)

	//initialize timer
	var timerEnd *time.Timer = time.NewTimer(0)
	<-timerEnd.C
	go timer(startTimerChan, timerEnd)

	//initialize sensors
	go elevio.PollButtons(drv_BtnChan)
	go elevio.PollFloorSensor(drv_NewFloorChan)
	go elevio.PollObstructionSwitch(drv_ObstrChan)
	go elevio.PollStopButton(drv_StopChan)

	//initialize network
	go network_buttonSender(drv_BtnChan, slaveToMasterOfflineBtnChan, ID)
	go network_broadcast(elevatorUpdateChan, slaveToMasterOfflineElevChan)
	go network_receiver(callsReceiverChan, masterToSlaveOfflineCallsChan, ID)

	//initialize fsm
	go fsm(ID, elevatorUpdateChan, callsReceiverChan, drv_NewFloorChan, drv_ObstrChan, drv_StopChan, startTimerChan, timerEnd)
}
