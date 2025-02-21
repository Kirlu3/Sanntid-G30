package master

import (
	"fmt"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	"github.com/Kirlu3/Sanntid-G30/heislab/slave"
	"github.com/mohae/deepcopy"
)

// it is important that this function doesnt block
func stateManager(initWorldview slave.WorldView, requestAssignment chan struct{}, slaveUpdate chan slave.EventMessage, backupUpdate chan []string,
	mergeState chan slave.WorldView, stateToBackup chan slave.WorldView, aliveBackups chan []string, requestBackupAck chan slave.Calls,
	stateToAssign chan slave.WorldView) {
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
					HallCalls: deepcopy.Copy(worldview.HallCalls).([config.N_ELEVATORS][config.N_FLOORS][config.N_BUTTONS - 1]bool),
					CabCalls:  deepcopy.Copy(worldview.HallCalls).([config.N_ELEVATORS][config.N_FLOORS]bool),
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

		case backups := <-backupUpdate:
			for i := range worldview.AliveElevators {
				worldview.AliveElevators[i] = false
			}
			for _, aliveIdx := range backups {
				worldview.AliveElevators[int(aliveIdx[0]-'0')] = true
			}
			stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)
			// maybe forward the update to receiveBackupAck on aliveBackups channel

		case otherMasterState := <-mergeState:
			fmt.Printf("otherMasterState: %v\n", otherMasterState)
			// inherit calls from otherMaster TODO
			stateToAssign <- deepcopy.Copy(worldview).(slave.WorldView)

		}
		stateToBackup <- deepcopy.Copy(worldview).(slave.WorldView)

	}
}
