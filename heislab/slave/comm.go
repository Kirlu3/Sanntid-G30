package slave

import (
	"encoding/gob"
	"fmt"
	"net"

	"github.com/Kirlu3/Sanntid-G30/heislab/driver-go/elevio"
)

type EventType int

const (
	Button       EventType = iota //In case of a button press or queue update, both ways
	Light                         //In case of a light update, only from master
	FloorArrival                  //In case of a floor arrival, only from slave
	Stuck                         //In case of a stuck elevator, only from slave
)

type EventMessage struct {
	Elevator Elevator           //Sends its own elevator struct, always
	Event    EventType          //Sends the type of event
	Btn      elevio.ButtonEvent //Sends a button in case of Button or Light
	Check    bool               //Sends a boolean for either Stuck or Light
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
