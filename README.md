# System overviev:

## Master 
Assgin calls to slave as well as sending them to the backup(s). Lights only turn on once the backups knows about the call ensuring service guarantee. 

## Backup 
A backup of the assigned calls in case master crashes 

## Slave  
Behaves according to the assigned calls from master