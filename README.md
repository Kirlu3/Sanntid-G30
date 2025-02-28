# Sanntid-G30
Sanntidsprosjekt gruppe 30 V25

TODO:
-Fix communication
-Add channel directions to *every* function
-Add comments to *every* function
    -Stuff like output and input and purpose

-Watchdog with program that restarts on a crash (would fix the occasional crashes on startup)

-Change config to include everything necessary, including: Base IP

-fix messages between slaveRx and assignOrders, backupAckRx

-fix graceful shutdown of master mode

-fix init of fsm (set ligths to 0 maybe?)

-how do the slaves handle losing their master?

-apparently go passes structs and arrays by value so deepcopy is probably not necessary?

-FIX: why do lights cab lights turn off on obstruction? something related to obstruction seems to break the elevator

-FIX: if an elevator has a single call in the direction it is coming from, it failt to clear it