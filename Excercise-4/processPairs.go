package main

import (
	"fmt"
	"net"
	"os/exec"
	"time"
)

func backup() {

}

func main() {

	fmt.Println("program started")
	cmd := exec.Command("gnome-terminal", "--", "go", "run", "/home/student/Documents/Sanntid-G30/Sanntid-G30/Excercise-4/processPairs.go")

	fmt.Println("Waiting for 5 seconds...")
	time.Sleep(5 * time.Second)
	fmt.Println("5 seconds have passed!")

	cmd.Run()

	// backup phase
	fmt.Println("starting backup phase")
	addr := ":20009"
	conn, err := net.ListenPacket("udp", addr)

	if err != nil {
		fmt.Println("Error starting")
		fmt.Println(err)
	}

	p :=make([]byte, 1024)
	conn.ReadFrom(p)

	n, _, client_err := conn.ReadFrom(p)

	p[0:n]
	if client_err != nil {
		fmt.Println("Client error")
		fmt.Println(client_err)
	}


	// primary phase
	for {
		fmt.Println("starting primary phase")

	}
}
