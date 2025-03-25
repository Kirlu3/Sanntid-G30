package config

const (
	N_FLOORS    = 4
	N_BUTTONS   = 3
	N_ELEVATORS = 3
)

const (
	WatchdogPort         = 15500
	MasterUpdatePort     = 30019
	BackupsUpdatePort    = 30029
	MasterBroadcastPort  = 30039
	BackupsBroadcastPort = 30049

	SlaveBroadcastPort = 40000
	SlaveButtonPort    = 40001
	SlaveAckPort       = 40002
	SlaveCallsPort     = 40003

	BackupBroadcastPeriodMs         = 1
	MasterBroadcastCallsPeriodMs    = 1
	MasterBroadcastAssignedPeriodMs = 10
	SlaveBroadcastPeriodMs          = 1

	ResendPeriodMs  = 10
	ResendTimeoutMs = 2000
)

const (
	DoorOpenDuration  = 3
	TimeBetweenFloors = 5
)
