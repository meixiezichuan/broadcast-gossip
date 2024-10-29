package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/meixiezichuan/broadcast-gossip/gossip"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

func runAgent(node string, port int, wg *sync.WaitGroup, ep int) {
	defer wg.Done()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {
		sig := <-sigs
		fmt.Println("Get sig ", sig, "Exiting ……")
		done <- true
	}()
	agent := gossip.InitAgent(node, port)
	fmt.Println(agent.NodeId, "Start Running ", ep, " epoch.")
	agent.Start(done, ep)
}

func main() {
	epstr := os.Args[1]
	ep, err := strconv.Atoi(epstr)
	if err != nil {
		ep = 10
	}

	// 模拟创建 5 个边缘节点
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {

		node := "node" + strconv.Itoa(i)
		port := 9898 + i
		//port := 9898
		wg.Add(1)
		go runAgent(node, port, &wg, ep)
	}

	// 等待所有节点完成任务
	wg.Wait()
	fmt.Println("All nodes have completed their tasks.")
}
