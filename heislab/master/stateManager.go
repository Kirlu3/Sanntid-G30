package master

import (
	"fmt"
	"strconv"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
	"github.com/mohae/deepcopy"
)

// it is important that this function doesnt block
func stateManager(initWorldview slave.WorldView, requestAssignment <-chan struct{}, slaveUpdate <-chan slave.EventMessage, backupUpdate <-chan []string,
	mergeState <-chan slave.WorldView, stateToBackup chan<- slave.WorldView, aliveBackupsCh <-chan []string, requestBackupAck chan<- slave.Calls,
	stateToAssign chan<- slave.WorldView, assignedRequests <-chan [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool, toSlaveCh chan<- [config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS]bool,
	endMasterPhase chan<- struct{}) {
	// aliveBackups might be redundant
	worldview := deepcopy.Copy(initWorldview).(slave.WorldView)
	for {
		fmt.Println("SM:New Loop")
		select {
		case assignments := <-assignedRequests: //updates state in worldview before sending out requests
			for elev := range config.N_ELEVATORS {
				worldview.Elevators[elev].Requests = assignments[elev]
			}
			toSlaveCh <- assignments
		case <-requestAssignment:
			fmt.Println("SM:reassignment")
			stateToAssign <- worldview

		case slaveMessage := <-slaveUpdate:
			println("SM:Received Slave Update")
			slaveId := slaveMessage.Elevator.ID
			switch slaveMessage.Event {

			case slave.Button:
				fmt.Println("SM:Button Event")
				if slaveMessage.Btn.Button == elevio.BT_Cab {
					worldview.CabCalls[slaveId][slaveMessage.Btn.Floor] = true
				} else {
					worldview.HallCalls[slaveMessage.Btn.Floor][slaveMessage.Btn.Button] = true
				}
				stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)
				requestBackupAck <- slave.Calls{
					HallCalls: deepcopy.Copy(worldview.HallCalls).([config.N_FLOORS][config.N_BUTTONS - 1]bool),
					CabCalls:  deepcopy.Copy(worldview.CabCalls).([config.N_ELEVATORS][config.N_FLOORS]bool),
				}

			case slave.FloorArrival:
				fmt.Println("SM: Floor Arrival Event")

				oldElevator := worldview.Elevators[slaveId]
				worldview.Elevators[slaveId] = slaveMessage.Elevator // i think it makes sense to update the whole state, again consider deepcopy
				newElevator := worldview.Elevators[slaveId]
				// should we reassign orders here?
				switch slaveMessage.Elevator.Behaviour {
				//If the elevator arrived at a floor and opened its door, it has cleared some unkown orders at that floor
				case slave.EB_DoorOpen:
					fmt.Println("SM: Clearing orders")
					//Updates cab orders:
					worldview.CabCalls[slaveId][newElevator.Floor] = newElevator.Requests[newElevator.Floor][elevio.BT_Cab]
					//Clears hall orders:
					for btn := range config.N_BUTTONS - 1 {
						//If the orders are different, prioritize the new ones, TODO: deal with if the elevator immediately cleared
						if oldElevator.Requests[newElevator.Floor][btn] != newElevator.Requests[newElevator.Floor][btn] {
							worldview.HallCalls[newElevator.Floor][btn] = newElevator.Requests[newElevator.Floor][btn]
						}
					}
					stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)
					requestBackupAck <- slave.Calls{
						HallCalls: deepcopy.Copy(worldview.HallCalls).([config.N_FLOORS][config.N_BUTTONS - 1]bool),
						CabCalls:  deepcopy.Copy(worldview.CabCalls).([config.N_ELEVATORS][config.N_FLOORS]bool),
					}
				}

			case slave.Stuck:
				fmt.Println("Stuck Event")

				worldview.Elevators[slaveId].Stuck = slaveMessage.Check
				stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)

			default:
				panic("invalid message event from slave")
			}

		case backups := <-aliveBackupsCh:
			for i := range worldview.AliveElevators {
				worldview.AliveElevators[i] = false
			}
			for _, aliveIdx := range backups {
				i, _ := strconv.Atoi(aliveIdx)
				worldview.AliveElevators[i] = true
			}
			stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)
			// maybe forward the update to receiveBackupAck on aliveBackups channel

		case otherMasterState := <-mergeState:
			fmt.Printf("otherMasterState: %v\n", otherMasterState)
			// inherit calls from otherMaster TODO
			if otherMasterState.OwnId > worldview.OwnId {

			} else if otherMasterState.OwnId < worldview.OwnId {
				if (isCallsSubset(slave.Calls{HallCalls: worldview.HallCalls, CabCalls: worldview.CabCalls},
					slave.Calls{HallCalls: otherMasterState.HallCalls, CabCalls: otherMasterState.CabCalls})) {
					endMasterPhase <- struct{}{}
				}
			}

			stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)

		}
		stateToBackup <- deepcopy.Copy(worldview).(slave.WorldView)
	}
}

// returns true if calls1 is a subset of calls2
func isCallsSubset(calls1 slave.Calls, calls2 slave.Calls) bool {
	for i := range config.N_ELEVATORS {
		for j := range config.N_FLOORS {
			if calls1.CabCalls[i][j] && !calls2.CabCalls[i][j] {
				return false
			}
		}
	}
	for i := range config.N_FLOORS {
		for j := range config.N_BUTTONS - 1 {
			if calls1.HallCalls[i][j] && !calls2.HallCalls[i][j] {
				return false
			}
		}
	}
	return true
}

// this is supposed to be some union function
func adhdhd(calls1 slave.Calls, calls2 slave.Calls) bool {
	for i := range config.N_ELEVATORS {
		for j := range config.N_FLOORS {
			calls1.CabCalls[i][j] = calls1.CabCalls[i][j] || calls2.CabCalls[i][j]
		}
	}
	for i := range config.N_FLOORS {
		for j := range config.N_BUTTONS - 1 {
			calls1.HallCalls[i][j] = calls1.HallCalls[i][j] || calls2.HallCalls[i][j]
		}
	}
	return true
}
