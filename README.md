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