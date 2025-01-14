package main

import (
	"fmt"
	"net"
)

func main() {
	addr := ":30000"
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		fmt.Println("Error starting")
		fmt.Println(err)
		return
	}
	q := []byte("Hello")
	server_addr := &net.UDPAddr{
		IP:   net.ParseIP("10.100.23.204"),
		Port: 20009,
	}
	conn.WriteTo(q, server_addr)

	p := make([]byte, 1024)
	n, _, client_err := conn.ReadFrom(p)
	if client_err != nil {
		fmt.Println("Client Error")
		fmt.Println(client_err)
		return
	}
	fmt.Println(string(p[0:n]))
}
