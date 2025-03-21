package config

const (
	N_FLOORS    = 4
	N_BUTTONS   = 3
	N_ELEVATORS = 3
)

// network
const (
	WatchdogPort         = 15500
	MasterUpdatePort     = 30019
	BackupsUpdatePort    = 30029
	MasterBroadcastPort  = 30039
	BackupsBroadcastPort = 30049
	SlaveBasePort        = 40000

	BackupMessagePeriodMs = 1
	MasterMessagePeriodMs = 1

	BroadcastMessagePeriodMs = 10

	ResendPeriodMs  = 10
	ResendTimeoutMs = 2000

	DoorOpenDuration  = 3
	TimeBetweenFloors = 5
)
