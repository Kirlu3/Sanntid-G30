package backup

import (
	"strconv"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/alive"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/bcast"
)

/*
# The backup routine listens to the master's UDP broadcasts and responds with the updated calls. If the backup loses connection with the master, it will transition to the master phase with its current list of calls.

Input: id

Returns: the calls that the backup has backed up
*/
func Run(id string) master.Calls {
	masterUpdateRxChan := make(chan alive.AliveUpdate)
	enableBackupTxChan := make(chan bool)
	backupCallsTxChan := make(chan struct {
		Calls master.Calls
		Id    int
	})
	masterCallsRxChan := make(chan struct {
		Calls master.Calls
		Id    int
	})

	go alive.Receiver(config.MasterUpdatePort, masterUpdateRxChan)
	go alive.Transmitter(config.BackupsUpdatePort, id, enableBackupTxChan)

	go bcast.Transmitter(config.BackupsBroadcastPort, backupCallsTxChan)
	go bcast.Receiver(config.MasterBroadcastPort, masterCallsRxChan)

	var masterUpdate alive.AliveUpdate
	var calls master.Calls

	idInt, err := strconv.Atoi(id)
	if err != nil {
		panic("backup received invalid id")
	}

	masterUpgradeCooldownTimer := time.NewTimer(5 * time.Second)

	for {
		select {
		case newCalls := <-masterCallsRxChan:
			if len(masterUpdate.Alive) > 0 && strconv.Itoa(newCalls.Id) == masterUpdate.Alive[0] {
				calls = newCalls.Calls
			}
		case masterUpdate = <-masterUpdateRxChan:

		case <-time.After(time.Millisecond * config.BackupBroadcastPeriodMs):
		}
		backupCallsTxChan <- master.BackupCalls{Calls: calls, Id: idInt}
		if len(masterUpdate.Alive) == 0 {
			select {
			case <-masterUpgradeCooldownTimer.C:
				enableBackupTxChan <- false
				return calls
			default:
			}
		}

	}
}
