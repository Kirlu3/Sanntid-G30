# Sanntid-G30
Sanntidsprosjekt gruppe 30 V25

Project is run by calling ./restart.sh elevator_id elevator_server_port


TODO:
-Add comments (docstring) to *every* function
    -Stuff like output and input and purpose

-Watchdog with program that restarts on a crash (would fix the occasional crashes on startup)

-FIX: Cab lights turning off when obstruction is flipped

-Consider: with a lot of packet loss so messages take a long time and new assignments are sent before a previous clear was received

-Consider: not using the id fields to do assingments

Ensure: That all cab orders from an offline elevator is sent to the corresponding master before it is potentially crashed by encountering another master with higher priority

-Implement single elevator mode for when network connection is missing

-Case of assignment not possible needs to be addressed, refuse assignment? but then the call is stored but no action is taken, which could lead to weird behaviour (we have the call in our system, but no lights, when is this update implemented) assign everything to the master? the exact assignments might not matter since no calls will actually be taken, but try to reduce inconsistency/ information loss