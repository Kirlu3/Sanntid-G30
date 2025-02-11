package Slave

import (
	"encoding/gob"
	"fmt"
	"net"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

type EventType int

const (
	Button EventType = iota
	FloorArrival
	Stuck
)

type EventMessage struct {
	Elev  Elevator
	Event EventType
	Btn   elevio.ButtonEvent
	Floor int
	Stuck bool
}

func Slavetcp(addr string, outgoing chan EventMessage, incoming chan EventMessage) {

	go sender(addr, outgoing)
	go receiver(addr, incoming)
	select {}
}

func sender(addr string, outgoing chan EventMessage) {
	conn, err := net.Dial("tcp", addr+":30000")
	if err != nil {
		fmt.Println(err) //what do we wanna do in this case?
		panic(err)
	}
	defer conn.Close()
	fmt.Println("Connected")
	enc := gob.NewEncoder(conn)
	for {
		msg := <-outgoing
		fmt.Println("sending message")
		enc.Encode(&msg)
	}
}

func receiver(addr string, incoming chan EventMessage) {
	listener, err := net.Listen("tcp", addr+":20000")
	if err != nil {
		fmt.Println(err) //what do we wanna do in this case?
		panic(err)
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err) //what do we wanna do in this case?
		panic(err)
	}
	defer conn.Close()

	dec := gob.NewDecoder(conn)
	for {
		var msg EventMessage
		dec.Decode(&msg)
		incoming <- msg
	}
}
