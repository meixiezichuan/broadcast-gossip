package gossip

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/meixiezichuan/broadcast-gossip/common"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
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
	NodeBuf       sync.Map
	Msgs          sync.Map
	Graph         *common.Graph
	MsgCnt        int
	BroadcastList []string
}

var TimeOutRev = 5

func InitAgent(nodeId string, port int) *Agent {
	agent := Agent{
		BroadcastAddr: "255.255.255.255:" + strconv.Itoa(port),
		ListenAddr:    ":" + strconv.Itoa(port),
		NodeId:        nodeId,
		Revision:      0,
		Graph:         common.NewGraph(),
		MsgCnt:        0,
	}
	return &agent
}

func (a *Agent) SetBroadcastList(l []string) {
	a.BroadcastList = l
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
	var sendMsgNodeId []string

	a.Msgs.Range(func(key, value interface{}) bool {
		n := key.(string)
		m := value.(HostMsg)
		paths := m.SendPaths

		for _, p := range paths {
			fmt.Println(a.NodeId, p, "exists in mlst")
			pn := p[len(p)-1]
			if true {
				s := common.SendMessage{
					PrevNode: pn,
					PrevAdj:  a.Graph.FindNeighbor(pn),
					NodeMsg:  m.Msg,
				}
				sendMsgs = append(sendMsgs, s)
				sendMsgNodeId = append(sendMsgNodeId, pn)
				break
			}
		}
		a.Msgs.Delete(n)
		return true
	})
	// Ensure the file is closed when done
	// add adj information
	for _, n := range a.Graph.FindNeighbor(a.NodeId) {
		if !common.Contains(sendMsgNodeId, n) {
			s := common.SendMessage{
				PrevNode: n,
				PrevAdj:  a.Graph.FindNeighbor(n),
				NodeMsg:  common.NodeMessage{},
			}
			sendMsgs = append(sendMsgs, s)
		}
	}
	sendMsg.Msgs = sendMsgs
	return sendMsg
}

func (a *Agent) Start(stopCh chan bool, ep int, distance int) {
	addr, err := net.ResolveUDPAddr("udp", a.ListenAddr)
	TimeOutRev = ep
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
	}()

	go a.BroadCast(stopCh, ep)
	a.ReceiveMsg(conn, stopCh, distance, ep)
}

func (a *Agent) Greeting() common.GossipMessage {
	dMsgs := []common.SendMessage{}
	for _, n := range a.Graph.FindNeighbor(a.NodeId) {
		var prevAdj []string
		for _, nn := range a.Graph.FindNeighbor(n) {
			prevAdj = append(prevAdj, nn)
		}
		sm := common.SendMessage{
			PrevNode: n,
			PrevAdj:  prevAdj,
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

func (a *Agent) DoBroadCast(msg common.GossipMessage, ep int) {
	l := 0
	if a.Revision < ep {
		l++
	}
	for _, m := range msg.Msgs {
		if !common.IsStructEmpty(m.NodeMsg) {
			if m.NodeMsg.Revision < ep {
				l++
			}
		}
	}
	a.MsgCnt = a.MsgCnt + l
	if len(a.BroadcastList) > 0 {
		for _, h := range a.BroadcastList {
			baddr := h + a.ListenAddr
			a.SendMsg(baddr, msg)
		}
	} else {
		baddr := a.BroadcastAddr
		a.SendMsg(baddr, msg)
	}
	fmt.Println(a.NodeId, "Send ", "msg: %v", msg)
}

func (a *Agent) BroadCast(stopCh chan bool, ep int) {
	time.Sleep(60 * time.Second)
	//t := rand.Intn(5)
	t := 20
	time.Sleep(time.Duration(t) * time.Second)
	fmt.Println(a.NodeId, " BroadCast")
	defer func() {
		fmt.Println(a.NodeId, "Sent Message Count: ", a.MsgCnt, " in ", a.Revision, "epochs")
		logPath := os.Getenv("LOG_PATH")
		if logPath == "" {
			logPath = "."
		}
		filename := logPath + "/" + a.NodeId
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			strings := fmt.Sprintf("Sent Message Count: %d in %d Epochs", a.MsgCnt, a.Revision)
			file.WriteString(strings)
		}
	}()
	for {
		select {
		case <-stopCh:
			fmt.Println("Received stop signal, stopping goroutine")
			return
		default:
			if a.Revision == ep+10 {
				fmt.Println("********", a.NodeId, "ran ", ep, " epoch finished.", "********")
				//stopCh <- true
				return
			}
			fmt.Println(a.NodeId, " in ", a.Revision, " graph1: ----")
			a.Graph.Display()
			a.UpdateGraph()
			fmt.Println(a.NodeId, " in ", a.Revision, " graph2: ----")
			a.Graph.Display()
			msg := a.generateGossipMessage()
			a.recordMsg(msg)
			a.DoBroadCast(msg, ep)
			a.Revision++
			time.Sleep(20 * time.Second)
		}
	}
}

func (a *Agent) recordMsg(msg common.GossipMessage) {
	// 打开或创建记录文件
	file, err := os.OpenFile("./gossip_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// 创建记录内容
	logContent := fmt.Sprintf("%s %d\n", msg.Self.NodeID, msg.Self.Revision)

	for _, m := range msg.Msgs {
		if m.NodeMsg.NodeID == "" {
			continue
		}
		logContent += fmt.Sprintf("%s %d\n", m.NodeMsg.NodeID, m.NodeMsg.Revision)
	}

	if _, err := file.WriteString(logContent); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}
}

func (a *Agent) isMsgRecorded(nodeID string, revision int) bool {
	// 打开日志文件
	file, err := os.Open("./gossip_logs.txt")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return false
	}
	defer file.Close()

	// 按行扫描文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var recordedNodeID string
		var recordedRevision int
		_, err := fmt.Sscanf(scanner.Text(), "%s %d", &recordedNodeID, &recordedRevision)
		if err != nil {
			continue
		}
		// 如果找到匹配项，返回 true
		if recordedNodeID == nodeID && recordedRevision == revision {
			return true
		}
	}
	// 如果读取文件时出错，打印错误
	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
	return false
}

func (a *Agent) UpdateGraph() {
	fmt.Println(a.NodeId, " NodeBuf: ", a.NodeBuf)
	a.NodeBuf.Range(func(key, value interface{}) bool {
		n := key.(string)
		r := value.(int)
		if a.Revision-r > TimeOutRev {
			a.Graph.RemoveEdge(a.NodeId, n)
		}
		return true
	})
}

func (a *Agent) WriteMsg(msg common.NodeMessage, ep int) {
	if msg.Revision >= ep || msg.Revision == 0 {
		return
	}
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "."
	}
	filename := logPath + "/" + a.NodeId
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening or creating file:", err)
	}
	defer file.Close()
	latency := a.Revision - msg.Revision
	msgWritten := msg.NodeID + "_" + strconv.Itoa(msg.Revision) + " " + strconv.Itoa(latency) + "\n"
	_, err = file.WriteString(msgWritten)
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func (a *Agent) SendMsg(baddr string, msg common.GossipMessage) {
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
	}
}

func (a *Agent) checkMsgSend(p []string) bool {
	// get local's prevNode
	prevNode := p[len(p)-1]

	// get subgraph in center of prevNode with hops 2
	subGraph := a.Graph.GetSubgraphWithinHops(prevNode, 2)

	allP := append(p, a.NodeId)
	if a.PathExistInMLST(subGraph, allP) {
		return true
	}
	return false
}
