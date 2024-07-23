package gossip

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

func (a *Agent) ReceiveMsg(conn *net.UDPConn, stopCh <-chan bool) {
	buf := make([]byte, 65535)
	for {
		select {
		case <-stopCh:
			fmt.Println("Received stop signal, stopping goroutine")
			return
		default:
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Failed to read UDP message: %v", err)
				continue
			}

			var msg GossipMessage
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}
			a.HandleMsg(msg)
		}
	}
}

func (a *Agent) HandleMsg(msg GossipMessage) {

	// add secNodeIP to Direct set
	srcNode := msg.Self.NodeID
	a.DirectSet = append(a.DirectSet, srcNode)
	a.Graph.AddEdge(a.NodeId, srcNode)
	for _, m := range msg.Direct {
		a.Graph.AddEdge(srcNode, m.NodeID)
		a.NodeBuf[m.NodeID] = append(a.NodeBuf[m.NodeID], srcNode)

		// TODO: compare m.Revision with already received am revision
		am, exist := a.Msgs[m.NodeID]
		if exist {
			if m.Revision <= am.Revision {
				continue
			}
		}
		a.Msgs[m.NodeID] = m
		// write to db
		n, err := a.DB.Get(m.NodeID)
		if err != nil {
			continue
		}
		if m.Revision > n.Revision {
			a.Write2DB(m)
		}
	}
}

// 处理接收到的Gossip消息
func (a *Agent) MessageNeedSend(msg NodeMessage) bool {
	msgNode := msg.NodeID
	mlst := a.Graph.FindMaxLeafTree(msgNode)
	if mlst.IsLeaf(a.NodeId) {
		return false
	}
	for _, sendNode := range a.NodeBuf[msgNode] {
		edge := []string{sendNode, a.NodeId}
		if mlst.PathExists(edge) {
			return true
		}
	}
	return false
}

func (a *Agent) Write2DB(msg NodeMessage) {
	a.DB.Lock()
	defer a.DB.Unlock()

	var existingValue string
	err := a.DB.db.QueryRow(`SELECT value FROM kv WHERE key = ?`, msg.NodeID).Scan(&existingValue)
	if err != nil || existingValue == "" {
		if err := a.DB.Set(msg.NodeID, msg); err == nil {
			log.Printf("Data synchronized: %s = %v\n", msg.NodeID, msg)
		}
	}
}
