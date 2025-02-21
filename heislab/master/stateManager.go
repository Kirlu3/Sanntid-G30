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
func stateManager(initWorldview slave.WorldView, requestAssignment chan struct{}, slaveUpdate chan slave.EventMessage, backupUpdate chan []string,
	mergeState chan slave.WorldView, stateToBackup chan slave.WorldView, aliveBackupsCh chan []string, requestBackupAck chan slave.Calls,
	stateToAssign chan slave.WorldView, endMasterPhase chan<- struct{}) {
	// aliveBackups might be redundant
	worldview := deepcopy.Copy(initWorldview).(slave.WorldView)
	for {
		select {
		case <-requestAssignment:
			stateToAssign <- worldview

		case slaveMessage := <-slaveUpdate:
			slaveId := slaveMessage.Elevator.ID
			switch slaveMessage.Event {

			case slave.Button:
				if slaveMessage.Btn.Button == elevio.BT_Cab {
					worldview.CabCalls[slaveId][slaveMessage.Btn.Floor] = slaveMessage.Check
				} else {
					worldview.HallCalls[slaveMessage.Btn.Floor][slaveMessage.Btn.Button] = slaveMessage.Check // do we have to be careful when removing order? i dont think so
				}
				stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)
				requestBackupAck <- slave.Calls{
					HallCalls: deepcopy.Copy(worldview.HallCalls).([config.N_FLOORS][config.N_BUTTONS - 1]bool),
					CabCalls:  deepcopy.Copy(worldview.CabCalls).([config.N_ELEVATORS][config.N_FLOORS]bool),
				}
				break

			case slave.FloorArrival:
				worldview.Elevators[slaveId] = slaveMessage.Elevator // i think it makes sense to update the whole state, again consider deepcopy
				// should we reassign orders here?
				switch slaveMessage.Elevator.Behaviour {
				//If the elevator arrived at a floor and opened its door, it has cleared orders
				case slave.EB_DoorOpen:

				}

				break

			case slave.Stuck:
				worldview.Elevators[slaveId].Stuck = slaveMessage.Check
				stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)
				break

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
			if (otherMasterState.OwnId < worldview.OwnId) {
				
			} else if (otherMasterState.OwnId > worldview.OwnId) {
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
	for i := 0; i < config.N_ELEVATORS; i++ {
		for j := 0; j < config.N_FLOORS; j++ {
			if calls1.CabCalls[i][j] && !calls2.CabCalls[i][j] {
				return false
			}
		}
	}
	for i := 0; i < config.N_FLOORS; i++ {
		for j := 0; j < config.N_BUTTONS-1; j++ {
			if calls1.HallCalls[i][j] && !calls2.HallCalls[i][j] {
				return false
			}
		}
	}
	return true
}

// this is supposed to be some union function
func adhdhd(calls1 slave.Calls, calls2 slave.Calls) bool {
	for i := 0; i < config.N_ELEVATORS; i++ {
		for j := 0; j < config.N_FLOORS; j++ {
			calls1.CabCalls[i][j] = calls1.CabCalls[i][j] || calls2.CabCalls[i][j]
		}
	}
	for i := 0; i < config.N_FLOORS; i++ {
		for j := 0; j < config.N_BUTTONS-1; j++ {
			calls1.HallCalls[i][j] = calls1.HallCalls[i][j] || calls2.HallCalls[i][j]
		}
	}
	return true
}