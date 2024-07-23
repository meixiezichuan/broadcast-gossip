package gossip

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

type NodeList []string

type Agent struct {
	BroadcastAddr string
	ListenAddr    string
	NodeId        string
	Revision      int
	DB            *Database
	DirectSet     NodeList
	LastDirectSet NodeList
	NodeBuf       map[string]NodeList
	Msgs          map[string]NodeMessage
	Graph         *Graph
}

func InitAgent(nodeId string) *Agent {
	agent := Agent{
		BroadcastAddr: "255.255.255.255:9898",
		ListenAddr:    ":9898",
		NodeId:        "node1",
		Revision:      0,
		DB:            InitDB(),
		DirectSet:     NodeList{},
		NodeBuf:       make(map[string]NodeList),
		Msgs:          make(map[string]NodeMessage),
	}
	return &agent
}

func InitDB() *Database {
	db, err := NewDatabase()
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	return db
}

// 生成Gossip消息
func (a *Agent) generateGossipMessage() GossipMessage {
	m := GossipMessage{}
	self := NodeMessage{a.NodeId, a.Revision, map[string]string{}}
	if a.Revision == 0 {
		m = a.Greeting()
	}
	dMsgs := []NodeMessage{}
	for _, d := range a.DirectSet {
		s, exists := a.Msgs[d]
		if exists {
			if a.MessageNeedSend(s) {
				dMsgs = append(dMsgs, s)
			}
			delete(a.Msgs, d)
		}
	}
	oMsgs := []NodeMessage{}
	for k, v := range a.Msgs {
		if a.MessageNeedSend(v) {
			oMsgs = append(oMsgs, v)
		}
		delete(a.Msgs, k)
	}
	m = GossipMessage{
		Self:   self,
		Direct: dMsgs,
		Other:  oMsgs,
	}
	a.NodeBuf = make(map[string]NodeList)
	return m
}

func (a *Agent) Start(stopCh <-chan bool) {
	addr, err := net.ResolveUDPAddr("udp", a.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP: %v", err)
	}
	defer conn.Close()

	go a.ReceiveMsg(conn, stopCh)

	a.BroadCast(stopCh)
}

func (a *Agent) Greeting() GossipMessage {
	dMsgs := []NodeMessage{}
	for _, d := range a.DirectSet {
		s, exists := a.Msgs[d]
		if exists {
			dMsgs = append(dMsgs, s)
		}
	}
	greeting := GossipMessage{
		Self: NodeMessage{
			NodeID:   a.NodeId,
			Revision: a.Revision,
		},
		Direct: dMsgs,
	}
	return greeting
}

func (a *Agent) DoBroadCast(msg GossipMessage) {
	a.Revision++
	addr, err := net.ResolveUDPAddr("udp", "255.255.255.255:9876")
	if err != nil {
		fmt.Printf("Error resolving address: %v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Printf("Error dialing UDP: %v\n", err)
		return
	}

	fmt.Printf("send msg: %v", msg)
	bytes, err := json.Marshal(msg)
	defer conn.Close()
	conn.Write(bytes)
}

func (a *Agent) BroadCast(stopCh <-chan bool) {
	for {
		select {
		case <-stopCh:
			fmt.Println("Received stop signal, stopping goroutine")
			return
		default:
			msg := a.generateGossipMessage()
			a.DoBroadCast(msg)
			time.Sleep(10 * time.Second)
		}
	}
}
