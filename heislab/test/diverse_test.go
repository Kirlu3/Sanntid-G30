package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/Kirlu3/Sanntid-G30/heislab/config"
	"github.com/Kirlu3/Sanntid-G30/heislab/master"
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

// monitor the master and backup update port
func TestListenBackupMasterUpdate(t *testing.T) {
	masterUpdateRxChan := make(chan peers.PeerUpdate)
	backupUpdateRxChan := make(chan peers.PeerUpdate)
	go peers.Receiver(config.MasterUpdatePort, masterUpdateRxChan)
	go peers.Receiver(config.BackupsUpdatePort, backupUpdateRxChan)

	for {
		select {
		case p := <-masterUpdateRxChan:
			fmt.Printf("master update:\n")
			fmt.Printf("  Masters:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case p := <-backupUpdateRxChan:
			fmt.Printf("backup update:\n")
			fmt.Printf("  Backups:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		}
	}
}

// monitor the master update port
func TestListenMasterUpdate(t *testing.T) {
	masterUpdateRxChan := make(chan peers.PeerUpdate)
	go peers.Receiver(config.MasterUpdatePort, masterUpdateRxChan)

	for {
		select {
		case p := <-masterUpdateRxChan:
			fmt.Printf("master update:\n")
			fmt.Printf("  Masters:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		}
	}
}

// monitor the backup update port
func TestListenBackupUpdate(t *testing.T) {
	backupUpdateRxChan := make(chan peers.PeerUpdate)
	go peers.Receiver(config.BackupsUpdatePort, backupUpdateRxChan)

	for {
		select {
		case p := <-backupUpdateRxChan:
			fmt.Printf("backup update:\n")
			fmt.Printf("  Backups:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		}
	}
}

func TestPeersEnableTx(t *testing.T) {
	id := "2"
	masterUpdateRxChan := make(chan peers.PeerUpdate)
	enableMasterTxChan := make(chan bool)
	go func() {
		for {
			select {
			case p := <-masterUpdateRxChan:
				fmt.Printf("master update:\n")
				fmt.Printf("  Masters:    %q\n", p.Peers)
				fmt.Printf("  New:      %q\n", p.New)
				fmt.Printf("  Lost:     %q\n", p.Lost)
				time.Sleep(time.Millisecond * 20)

			}
		}
	}()
	go peers.Transmitter(config.MasterUpdatePort, id, enableMasterTxChan)
	go peers.Receiver(config.MasterUpdatePort, masterUpdateRxChan)
	// time.Sleep(time.Millisecond*10)
	// masterTxEnable <- false

	enableMasterTxChan2 := make(chan bool)
	go peers.Transmitter(config.MasterUpdatePort, "3", enableMasterTxChan2)
	enable := false
	for {
		time.Sleep(time.Millisecond * 2000)
		enable = !enable
		enableMasterTxChan2 <- enable
	}

	// time.Sleep(time.Second*30)

}

func TestCtx(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			println("do work 1")
			time.Sleep(time.Second)
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			println("do work 2")
			time.Sleep(time.Second)
		}
	}()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			println("do work 3")
			time.Sleep(time.Second)
		}
	}()

	time.Sleep(time.Second * 10)
	cancel()
	time.Sleep(time.Second * 10)

}

func TestCompareStructs(t *testing.T) {
	var calls1 master.BackupCalls
	var calls2 master.BackupCalls
	calls1.Calls.CabCalls[0][0] = false
	fmt.Println(calls1 == calls2)
}

func TestChans(t *testing.T) {
	backupCallsChan := make(chan struct {
		Calls master.Calls
		Id    int
	})

	assignCallsChan := make(chan struct {
		Calls          master.Calls
		AliveElevators [config.N_ELEVATORS]bool
	})

	updateCallsChan := make(chan struct {
		Calls   master.Calls
		AddCall bool
	})

	// aChan := make(chan master.UpdateCalls)
	func(chan struct {
		Calls          master.Calls
		AliveElevators [config.N_ELEVATORS]bool
	}) {
		return
	}(assignCallsChan)
	var a master.UpdateCalls
	updateCallsChan <- a

	var calls master.Calls
	id := 1
	var s master.BackupCalls

	backupCallsChan <- master.BackupCalls{Calls: calls, Id: id}
	backupCallsChan <- s

}
