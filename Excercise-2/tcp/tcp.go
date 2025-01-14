package main

import (
	"fmt"
	"net"
)

func receive(conn net.Conn) {
	defer conn.Close()
	for {
		p := make([]byte, 1024)
		n, _ := conn.Read(p)
		fmt.Println(string(p[0:n]))
	}
}

func send(conn net.Conn) {
	defer conn.Close()
	msg := []byte("Connect to: 10.100.23.19:20009\x00")
	conn.Write(msg)

}

func send2(conn net.Conn) {
	for {
		msg := []byte("Hi\x00")
		conn.Write(msg)
	}
}

func main() {
	addr := "10.100.23.204:33546"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("Error starting", err)
		return
	}
	send(conn)

	addr_2 := "10.100.23.19:20009"
	listener, err_2 := net.Listen("tcp", addr_2)

	fmt.Println("listening")

	if err_2 != nil {
		fmt.Println("Error starting", err_2)
		return
	}

	defer listener.Close()

	conn_2, listen_err := listener.Accept()

	if listen_err != nil {
		fmt.Println("Error accepting connection", listen_err)
		return
	}

	fmt.Println("accepted connection")
	go receive(conn_2)
	go send2(conn_2)

	select {}
}
