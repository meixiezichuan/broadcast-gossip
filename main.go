package main

import (
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/meixiezichuan/broadcast-gossip/gossip"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

func runSimulation(node string, port int, wg *sync.WaitGroup, distance int, ep int) {
	defer wg.Done()
	runAgent(node, port, distance, ep)
}

func runAgent(node string, port int, distance int, ep int) {
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
	agent.Start(done, ep, distance)
}

func Simulation(ep int) {
	// 模拟创建 5 个边缘节点
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {

		node := "node" + strconv.Itoa(i)
		port := 9898 + i
		//port := 9898
		wg.Add(1)
		go runSimulation(node, port, &wg, 3, ep)
	}

	// 等待所有节点完成任务
	wg.Wait()
}

func getLocalIP() string {
	i := rand.Intn(100)
	ip := "node" + strconv.Itoa(i)
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Get Interface Error:", err)
		return ""
	}

	for _, addr := range addrs {
		// Check if the address is an IP address (skip loopback)
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				fmt.Println("Local IP address:", ip)
				return ip
			}
		}
	}
	return ip
}

func main() {
	ep := 100
	distance := 100
	if len(os.Args) > 1 {
		epstr := os.Args[1]
		e, err := strconv.Atoi(epstr)
		if err == nil {
			ep = e
		}

		lenstr := os.Args[2]
		e, err = strconv.Atoi(lenstr)
		if err == nil {
			distance = e
		}
	}

	node, exist := os.LookupEnv("Hostname")
	if !exist {
		node = getLocalIP()
	}

	port := 9898
	strport, exist := os.LookupEnv("BroadcastPort")
	if exist {
		port, _ = strconv.Atoi(strport)
	}
	runAgent(node, port, distance, ep)
	fmt.Println("All nodes have completed their tasks.")
}
