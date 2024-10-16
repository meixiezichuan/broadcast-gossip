package gossip

import (
	"encoding/json"
	"fmt"
	"github.com/meixiezichuan/broadcast-gossip/common"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type NodeList []string

type Path []string

type HostMsg struct {
	Msg       common.NodeMessage
	SendPaths []Path
}

type NodeInfo struct {
	Cpu     string
	Battery string
	Mem     string
}

type Agent struct {
	BroadcastAddr string
	ListenAddr    string
	NodeId        string
	Revision      int
	DB            *common.Database
	NodeBuf       map[string]int
	Msgs          map[string]HostMsg
	Graph         *common.Graph
}

var TimeOutRev = 5

func InitAgent(nodeId string, port int) *Agent {
	agent := Agent{
		BroadcastAddr: "255.255.255.255:" + strconv.Itoa(port),
		ListenAddr:    ":" + strconv.Itoa(port),
		NodeId:        nodeId,
		Revision:      0,
		DB:            InitDB(nodeId),
		NodeBuf:       make(map[string]int),
		Msgs:          make(map[string]HostMsg),
		Graph:         common.NewGraph(),
	}
	return &agent
}

func InitDB(name string) *common.Database {
	db, err := common.NewDatabase(name)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return db
}

// 生成Gossip消息
func (a *Agent) generateGossipMessage() common.GossipMessage {
	sendMsg := common.GossipMessage{}
	self := common.NodeMessage{a.NodeId, a.Revision, map[string]string{}}
	if a.Revision == 0 {
		sendMsg = a.Greeting()
		return sendMsg
	}

	self.Data = common.GenerateNodeInfo()
	sendMsg.Self = self
	var sendMsgs []common.SendMessage
	for n, m := range a.Msgs {
		paths := m.SendPaths
		for _, p := range paths {
			allP := append(p, a.NodeId)
			if a.PathExistInMLST(allP) {
				fmt.Println(a.NodeId, allP, "exists in mlst")
				s := common.SendMessage{
					PrevNode: p[len(p)-1],
					NodeMsg:  m.Msg,
				}
				sendMsgs = append(sendMsgs, s)
				break
			}
		}
		delete(a.Msgs, n)
	}
	sendMsg.Msgs = sendMsgs
	return sendMsg
}

func (a *Agent) Start(stopCh <-chan bool) {
	addr, err := net.ResolveUDPAddr("udp", a.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}
	fmt.Println(a.NodeId, " udpListen: ", addr)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("%s Failed to listen on UDP: %v", a.NodeId, err)
	}
	defer conn.Close()

	go a.ReceiveMsg(conn, stopCh)
	t := rand.Intn(5)
	time.Sleep(time.Duration(t) * time.Second)
	a.BroadCast(stopCh)
}

func (a *Agent) Greeting() common.GossipMessage {
	dMsgs := []common.SendMessage{}
	for _, n := range a.Graph.FindNeighbor(a.NodeId) {
		sm := common.SendMessage{
			PrevNode: n,
		}
		dMsgs = append(dMsgs, sm)
	}

	greeting := common.GossipMessage{
		Self: common.NodeMessage{
			NodeID:   a.NodeId,
			Revision: a.Revision,
		},
		Msgs: dMsgs,
	}
	return greeting
}

func (a *Agent) DoBroadCast(msg common.GossipMessage) {
	for i := 0; i < 5; i++ {
		baddr := "255.255.255.255:" + strconv.Itoa(9898+i)
		addr, err := net.ResolveUDPAddr("udp", baddr)
		if err != nil {
			fmt.Printf("%s Error resolving address: %v\n", a.NodeId, err)
			return
		}

		conn, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			fmt.Printf("%s Error dialing UDP: %v\n", a.NodeId, err)
			return
		}

		bytes, err := json.Marshal(msg)
		defer conn.Close()
		_, err = conn.Write(bytes)
		if err != nil {
			fmt.Printf("%s Error write UDP: %v\n", a.NodeId, err)
			return
		}
	}

	fmt.Println(a.NodeId, "Send ", "msg: %v", msg)
}

func (a *Agent) BroadCast(stopCh <-chan bool) {
	fmt.Println(a.NodeId, " BroadCast")
	for {
		select {
		case <-stopCh:
			fmt.Println("Received stop signal, stopping goroutine")
			return
		default:
			fmt.Println(a.NodeId, " graph1: ----")
			a.Graph.Display()
			a.UpdateGraph()
			fmt.Println(a.NodeId, " graph2: ----")
			a.Graph.Display()
			msg := a.generateGossipMessage()
			a.DoBroadCast(msg)
			a.Revision++
			time.Sleep(5 * time.Second)
		}
	}
}

func (a *Agent) UpdateGraph() {
	fmt.Println(a.NodeId, " NodeBuf: ", a.NodeBuf)
	for n, r := range a.NodeBuf {
		if a.Revision-r > TimeOutRev {
			a.Graph.RemoveEdge(a.NodeId, n)
		}
	}
}
