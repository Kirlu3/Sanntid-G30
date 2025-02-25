package test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/network/peers"
)

func TestNetwork(t *testing.T) {
	// network()
}

func TestFoo(t *testing.T) {
	foo()
}

func foo() {
	var a []string
	s := "one"
	num, err := strconv.Atoi(s)
	fmt.Printf("num: %v\n", num)
	fmt.Printf("err: %v\n", err)

	for i, val := range a {
		fmt.Printf("a[i]: %v\n", a[i])
		fmt.Printf("val: %v\n", val)
	}

}

func TestListenBackupMasterUpdate(t *testing.T) {
	masterUpdateCh := make(chan peers.PeerUpdate)
	backupUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(config.MasterUpdatePort, masterUpdateCh)
	go peers.Receiver(config.BackupsUpdatePort, backupUpdateCh)

	for {
		time.Sleep(100 * time.Millisecond)
		select {
		case p := <-masterUpdateCh:
			fmt.Printf("master update:\n")
			fmt.Printf("  Masters:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case p := <-backupUpdateCh:
			fmt.Printf("backup update:\n")
			fmt.Printf("  Backups:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		}
	}
}

func TestListenMasterUpdate(t *testing.T) {
	masterUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(config.MasterUpdatePort, masterUpdateCh)

	for {
		select {
		case p := <-masterUpdateCh:
			fmt.Printf("master update:\n")
			fmt.Printf("  Masters:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		}
	}
}

func TestListenBackupUpdate(t *testing.T) {
	backupUpdateCh := make(chan peers.PeerUpdate)
	go peers.Receiver(config.BackupsUpdatePort, backupUpdateCh)

	for {
		time.Sleep(1 * time.Second)
		select {
		case p := <-backupUpdateCh:
			fmt.Printf("backup update:\n")
			fmt.Printf("  Backups:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		}
	}
}
