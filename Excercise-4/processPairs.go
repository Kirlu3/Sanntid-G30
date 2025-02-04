package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os/exec"
	"time"
)

func timer(t_start chan bool, t_end *time.Timer, quit chan bool, timeout time.Duration) {
	for {
		select {
		case a := <-t_start:
			if a {
				t_end.Reset(timeout)
			}
		case <-quit:
			return
		}
	}
}

func read(conn net.PacketConn, msg chan int) {
	p := make([]byte, 1024)
	for {
		_, _, err := conn.ReadFrom(p)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		msg <- int(p[0])
	}
}

func main() {
	var t_end *time.Timer = time.NewTimer(0)
	<-t_end.C
	t_start := make(chan bool)

	quit1 := make(chan bool)

	var num int

	fmt.Println("Program started")

	// backup phase
	fmt.Println("Starting connection")
	conn, err := net.ListenPacket("udp", ":2000")

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	defer conn.Close()

	msg := make(chan int)
	go read(conn, msg)

	//actual backup phase
	fmt.Println("Starting backup")
	go timer(t_start, t_end, quit1, time.Second*3)
	t_start <- true
backup:
	for {
		select {
		case <-t_end.C:
			fmt.Println("Backup timeout")
			quit1 <- true
			break backup
		case num = <-msg:
			continue
		}
	}

	// primary phase
	fmt.Println("Starting primary phase")

	cmd := exec.Command("gnome-terminal", "--", "go", "run", "processPairs.go")
	err = cmd.Start()

	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	addr, err := net.ResolveUDPAddr("udp4", "localhost:2000")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	go timer(t_start, t_end, quit1, time.Second*5)
	t_start <- true
	for {
		time.Sleep(500 * time.Millisecond)
		select {
		case <-t_end.C:
			return
		default:
			num = num + 1
			buf := new(bytes.Buffer)
			binary.Write(buf, binary.LittleEndian, int32(num))
			_, err = conn.WriteTo(buf.Bytes(), addr)
			if err != nil {
				fmt.Println(err)
				panic(err)
			}
			fmt.Println(num)
		}
	}
}
