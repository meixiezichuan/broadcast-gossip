package gossip

import (
	"encoding/json"
	"fmt"
	"github.com/meixiezichuan/broadcast-gossip/common"
	"log"
	"net"
)

func (a *Agent) ReceiveMsg(conn *net.UDPConn, stopCh <-chan bool) {
	fmt.Println(a.NodeId, " receive msg ")
	buf := make([]byte, 65535)
	for {
		select {
		case <-stopCh:
			fmt.Println(a.NodeId, "Received stop signal, stopping goroutine")
			return
		default:
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				//log.Printf("%s Failed to read UDP message: %v", a.NodeId, err)
				continue
			}
			fmt.Println(a.NodeId, " receive msg n: ", n)
			var msg common.GossipMessage
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			a.HandleMsg(msg)
		}
	}
}

func (a *Agent) HandleMsg(msg common.GossipMessage) {
	fmt.Println(a.NodeId, "handle ", msg)
	//1. first get network topo
	// get direct node msg
	dmsg := msg.Self
	if dmsg.NodeID == a.NodeId {
		return
	}
	// 加入一跳桶
	a.WriteMsg(dmsg)
	a.Graph.AddEdge(a.NodeId, dmsg.NodeID)

	rev, exist := a.NodeBuf[dmsg.NodeID]
	if exist {
		if rev < dmsg.Revision {
			a.NodeBuf[dmsg.NodeID] = dmsg.Revision
		}
	} else {
		a.NodeBuf[dmsg.NodeID] = dmsg.Revision
	}

	// add msg
	path := Path{dmsg.NodeID}
	a.UpdateMsgs(dmsg, path)

	// handle other msg
	for _, m := range msg.Msgs {
		a.Graph.AddEdge(dmsg.NodeID, m.PrevNode)
		// handle msg
		if !common.IsStructEmpty(m.NodeMsg) {
			a.WriteMsg(m.NodeMsg)
			path = Path{m.PrevNode, dmsg.NodeID}
			if m.NodeMsg.NodeID != a.NodeId {
				a.UpdateMsgs(m.NodeMsg, path)
			}
		}
	}
}

// 处理接收到的Gossip消息
func (a *Agent) PathExistInMLST(p Path) bool {

	preNode := p[0]
	mlst, _ := a.Graph.MLST10(preNode)
	fmt.Println(a.NodeId, " root: ", preNode, " path: ", p, " mlst: ")
	mlst.Display()
	// if node is leaf, return false
	if mlst.IsLeaf(a.NodeId) {
		return false
	}

	if mlst.PathExists(p) {
		return true
	}
	return false
}

func (a *Agent) UpdateMsgs(msg common.NodeMessage, path Path) {
	_, exist := a.Msgs[msg.NodeID]
	Hm := HostMsg{
		Msg: msg,
	}
	if exist {
		sendpath := append(a.Msgs[msg.NodeID].SendPaths, path)
		Hm.SendPaths = sendpath
	} else {
		Hm.SendPaths = []Path{path}
	}
	a.Msgs[msg.NodeID] = Hm
}
