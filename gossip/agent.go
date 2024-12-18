package gossip

import (
	"encoding/json"
	"fmt"
	"github.com/meixiezichuan/broadcast-gossip/common"
	"log"
	"math/rand"
	"net"
	"os"
	"sort"
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
	MsgCnt        int
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
		MsgCnt:        0,
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
	filename := a.NodeId
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file:", err)
	}

	msgWritten := self.NodeID + "_" + strconv.Itoa(self.Revision) + "\n"
	_, err = file.WriteString(msgWritten)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
	//a.Write2DB(self)

	self.Data = common.GenerateNodeInfo()
	sendMsg.Self = self
	var sendMsgs []common.SendMessage
	var sendMsgNodeId []string
	for n, m := range a.Msgs {
		//s := common.SendMessage{
		//	PrevNode: n,
		//	NodeMsg:  m.Msg,
		//}
		//sendMsgs = append(sendMsgs, s)
		//a.Write2DB(m.Msg)
		msgWritten = m.Msg.NodeID + "_" + strconv.Itoa(m.Msg.Revision) + "\n"
		_, err = file.WriteString(msgWritten)
		if err != nil {
			fmt.Println("Error writing to file:", err)
		}
		paths := m.SendPaths
		sort.Slice(paths, func(i, j int) bool {
			return paths[i][0] < paths[j][0]
		})
		for _, p := range paths {
			allP := append(p, a.NodeId)
			if a.PathExistInMLST(allP) {
				fmt.Println(a.NodeId, allP, "exists in mlst")
				s := common.SendMessage{
					PrevNode: p[len(p)-1],
					NodeMsg:  m.Msg,
				}
				sendMsgs = append(sendMsgs, s)
				sendMsgNodeId = append(sendMsgNodeId, m.Msg.NodeID)
				break
			}
		}
		delete(a.Msgs, n)
	}
	// Ensure the file is closed when done
	file.Close()
	// add adj information
	for n, v := range a.NodeBuf {
		// check if timeout
		if a.Revision-v > TimeOutRev {
			continue
		}
		if !common.Contains(sendMsgNodeId, n) {
			s := common.SendMessage{
				PrevNode: n,
				NodeMsg:  common.NodeMessage{},
			}
			sendMsgs = append(sendMsgs, s)
		}
	}
	sendMsg.Msgs = sendMsgs
	return sendMsg
}

func (a *Agent) Start(stopCh chan bool, ep int) {
	addr, err := net.ResolveUDPAddr("udp", a.ListenAddr)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}
	fmt.Println(a.NodeId, " udpListen: ", addr)
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("%s Failed to listen on UDP: %v", a.NodeId, err)
	}
	defer func() {
		conn.Close()
		fmt.Println(a.NodeId, "Sent Message Count: ", a.MsgCnt, " in ", a.Revision, "epochs")
	}()

	go a.ReceiveMsg(conn, stopCh)
	t := rand.Intn(5)
	time.Sleep(time.Duration(t) * time.Second)
	a.BroadCast(stopCh, ep)
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
	l := 1
	for _, m := range msg.Msgs {
		if !common.IsStructEmpty(m.NodeMsg) {
			l++
		}
	}
	a.MsgCnt = a.MsgCnt + l
	//for i := 0; i < 5; i++ {
	//	baddr := "255.255.255.255:" + strconv.Itoa(9898+i)
	baddr := a.BroadcastAddr
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
	//}

	fmt.Println(a.NodeId, "Send ", "msg: %v", msg)
}

func (a *Agent) BroadCast(stopCh chan bool, ep int) {
	fmt.Println(a.NodeId, " BroadCast")
	for {
		select {
		case <-stopCh:
			fmt.Println("Received stop signal, stopping goroutine")
			return
		default:
			if a.Revision == ep {
				fmt.Println("********", a.NodeId, "ran ", ep, " epoch finished.", "********")
				stopCh <- true
				return
			}
			fmt.Println(a.NodeId, " in ", a.Revision, " graph1: ----")
			a.Graph.Display()
			a.UpdateGraph()
			fmt.Println(a.NodeId, " in ", a.Revision, " graph2: ----")
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
