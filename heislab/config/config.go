package config

const (
	NumFloors    = 4
	NumBtns      = 3
	NumElevators = 3
)

const (
	MasterUpdatePort     = 30019
	BackupsUpdatePort    = 30029
	MasterBroadcastPort  = 30039
	BackupsBroadcastPort = 30049

	SlaveBroadcastPort = 40000
	SlaveButtonPort    = 40001
	SlaveAckPort       = 40002
	SlaveCallsPort     = 40003
)

const (
	BackupBroadcastPeriodMs         = 1
	MasterBroadcastCallsPeriodMs    = 1
	MasterBroadcastAssignedPeriodMs = 10
	SlaveBroadcastPeriodMs          = 1

	ResendPeriodMs  = 10
	ResendTimeoutMs = 2000

	CheckBackupAckMs = 50
)

const (
	DoorOpenDuration  = 3
	TimeBetweenFloors = 5
)
