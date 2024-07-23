package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/meixiezichuan/broadcast-gossip/gossip"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		fmt.Println("Get sig ", sig, "Exiting ……")
		done <- true
	}()
	agent := gossip.InitAgent("node1")
	agent.Start(done)
}
