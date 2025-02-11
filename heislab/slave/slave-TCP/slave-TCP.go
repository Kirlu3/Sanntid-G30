package Slavetcp

import (
	"encoding/gob"
	"fmt"
	"net"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
	Slave "github.com/Kirlu3/Sanntid-G30/heislab/slave"
)

type EventType int

const (
	Button EventType = iota
	FloorArrival
	Stuck
)

type EventMessage struct {
	elevator Slave.Elevator
	event    EventType
	btn      elevio.ButtonEvent
	floor    int
	stuck    bool
}

func Slavetcp(addr string, outgoing chan EventMessage, in chan EventMessage) {

	s_conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err) //what do we wanna do in this case?
		panic(err)
	}
	defer s_conn.Close()

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println(err) //what do we wanna do in this case?
		panic(err)
	}
	defer listener.Close()

	r_conn, err := listener.Accept()
	if err != nil {
		fmt.Println(err) //what do we wanna do in this case?
		panic(err)
	}
	defer r_conn.Close()

	incoming := make(chan EventMessage)

	go sender(s_conn, outgoing)

	go receiver(r_conn, incoming)

	select {}
}

func sender(conn net.Conn, outgoing chan EventMessage) {
	for {
		msg := <-outgoing
		enc := gob.NewEncoder(conn)
		enc.Encode(msg)
	}
}

func receiver(conn net.Conn, incoming chan EventMessage) {
	dec := gob.NewDecoder(conn)
	for {
		var msg EventMessage
		dec.Decode(&msg)
		incoming <- msg
	}
}
