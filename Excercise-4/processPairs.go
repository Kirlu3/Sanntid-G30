package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// func main() {

// 	addr := ":20009"
// 	conn, err := net.ListenPacket("udp", addr)

// 	fmt.Println("program started")
// 	cmd := exec.Command("gnome-terminal", "--", "go", "run", "/home/student/Documents/Sanntid-G30/Sanntid-G30/Excercise-4/processPairs.go")

// 	fmt.Println("Waiting for 5 seconds...")
// 	time.Sleep(5 * time.Second)
// 	fmt.Println("5 seconds have passed!")

// 	cmd.Run()

// 	// backup phase
// 	fmt.Println("starting backup phase")

// 	if err != nil {
// 		fmt.Println("Error starting")
// 		fmt.Println(err)
// 	}

// 	p :=make([]byte, 1024)
// 	conn.ReadFrom(p)

// 	n, _, client_err := conn.ReadFrom(p)

// 	p[0:n]
// 	if client_err != nil {
// 		fmt.Println("Client error")
// 		fmt.Println(client_err)
// 	}

// 	// primary phase
// 	for {
// 		fmt.Println("starting primary phase")

// 	}
// }

func main() {

	// backup phase
	fmt.Println("starting backup phase")

	addr := ":20009"
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("error connecting: ", err)
		return
	}

	fmt.Println("Listening on", addr)

	// defer conn.Close()

	receivedData := make([]byte, 4)
	var receivedInt int32

	for {
		conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		_, _, err := conn.ReadFromUDP(receivedData)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				// Timeout occurred (no message received in 5 seconds)
				fmt.Println("No message received for 5 seconds. end packup phase")
				break
			}
		}
		// update the count
		readBuf := bytes.NewBuffer(receivedData)
		binary.Read(readBuf, binary.BigEndian, &receivedInt)
		print("received int: ", receivedInt)
	}

	conn.Close()

	// primary phase
	fmt.Println("starting primary phase...")

	primaryAddr := ":30009"
	primaryUDPAddr, _ := net.ResolveUDPAddr("udp", primaryAddr)

	primaryConn, err := net.ListenUDP("udp", primaryUDPAddr)
	if err != nil {
		fmt.Println("primary error connecting: ", err)
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	dir := filepath.Dir(execPath)
	println("directory: ", dir)
	println("execpath: ", execPath)

	fmt.Println("program started")
	cmd := exec.Command("gnome-terminal", "--", "bash", "-c", execPath+"; exec bash")

	fmt.Println("Waiting for 5 seconds...")
	time.Sleep(2 * time.Second)
	fmt.Println("5 seconds have passed!")

	cmd.Run()

	startNum := receivedInt + 1
	for j := startNum; true; j++ {
		var buf bytes.Buffer
		err = binary.Write(&buf, binary.BigEndian, int32(j))
		if err != nil {
			fmt.Println("Error converting int to byte slice:", err)
			return
		}

		data := buf.Bytes()
		primaryConn.WriteTo(data, udpAddr)
		fmt.Println(j)
		time.Sleep(2 * time.Second)
	}
}
