package main

import "Heislab/Driver-go/elevio"
import "fmt"
import "Heislab/fsm"

func main(){

    numFloors := elevio.NumFloors

    elevio.Init("localhost:15657", numFloors)
    
    //var d elevio.MotorDirection = elevio.MD_Up
    //elevio.SetMotorDirection(d)
    
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)
    t_start     := make(chan bool)
    t_end       := make(chan bool)

    
    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
    go elevio.PollStopButton(drv_stop)
    go fms.timer(t_start, t_end)
    
    for {
        select {
        case a := <- drv_buttons:
            fsm.onRequestButtonPress(a, t_start)

        case a := <- drv_floors:
            fsm.onFloorArrival(a)

        case a := <- drv_obstr:
            fsm.onObstruction(a)

        case a := <- drv_stop:
            fsm.onStopButtonPress(a)
        }
        case a := <- t_end:
            fsm.onTimerEnd(a)
    }    
}