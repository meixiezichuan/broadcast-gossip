package gossip

import (
	"encoding/json"
	"fmt"
	"github.com/meixiezichuan/broadcast-gossip/common"
	"log"
	"net"
)

func (a *Agent) ReceiveMsg(conn *net.UDPConn, stopCh <-chan bool, distance int, ep int) {
	buf := make([]byte, 65507)

	// 创建一个 map，key 为 "NodeID-Revision" 作为唯一标识符，value 为分片 map
	fragmentMaps := make(map[string]map[int]common.GossipMessageWithChunks)
	totalChunks := make(map[string]int) // 保存每个 NodeID+Revision 对应的总分片数

	for {
		select {
		case <-stopCh:
			fmt.Println(a.NodeId, "Received stop signal, stopping goroutine")
			return
		default:
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				continue
			}

			// 反序列化消息
			var msg common.GossipMessageWithChunks
			if err := json.Unmarshal(buf[:n], &msg); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// 构造唯一标识符 "NodeID-Revision"
			uniqueKey := fmt.Sprintf("%s-%d", msg.NodeID, msg.Revision)

			// 如果还没有为该唯一标识符创建 map，需要创建它
			if _, exists := fragmentMaps[uniqueKey]; !exists {
				fragmentMaps[uniqueKey] = make(map[int]common.GossipMessageWithChunks)
			}

			// 将分片存储在该唯一标识符的 map 中，key 为分片索引
			fragmentMaps[uniqueKey][msg.ChunkIndex] = msg

			// 如果这是第一次接收该消息的分片，记录总的分片数
			if totalChunks[uniqueKey] == 0 {
				totalChunks[uniqueKey] = msg.TotalChunks
			}

			// 如果接收到所有分片，重新组装数据
			if len(fragmentMaps[uniqueKey]) == totalChunks[uniqueKey] {
				var fullData []byte
				for i := 0; i < totalChunks[uniqueKey]; i++ {
					if fragment, ok := fragmentMaps[uniqueKey][i]; ok {
						fullData = append(fullData, fragment.Data...)
					}
				}

				// 组装完成，处理完整消息
				var fullMsg common.GossipMessage
				if err := json.Unmarshal(fullData, &fullMsg); err != nil {
					log.Printf("Failed to unmarshal full message: %v", err)
				} else {
					// 处理完整的 GossipMessage
					a.HandleMsg(fullMsg, distance, ep)
				}

				// 清空当前节点的分片数据，准备接收下一条消息
				delete(fragmentMaps, uniqueKey)
				delete(totalChunks, uniqueKey)
			}
		}
	}
}

func (a *Agent) HandleMsg(msg common.GossipMessage, distance int, ep int) {

	//1. first get network topo
	// get direct node msg
	dmsg := msg.Self
	parentIP := common.Ip2int(net.ParseIP(dmsg.NodeID))
	localIP := common.Ip2int(net.ParseIP(a.NodeId))

	if dmsg.NodeID == a.NodeId {
		return
	}

	//lastDot := strings.LastIndex(dmsg.NodeID, ".")
	//lastPart := dmsg.NodeID[lastDot+1:]
	//receiveNum, _ := strconv.Atoi(lastPart)
	//
	//lastDot = strings.LastIndex(a.NodeId, ".")
	//lastPart = a.NodeId[lastDot+1:]
	//num, _ := strconv.Atoi(lastPart)

	//distance1 := (receiveNum - num + 255) % 255
	//distance2 := (num - receiveNum + 255) % 255
	//if distance1 > distance && distance2 > distance {
	//	return
	//}
	fmt.Printf("localIP: %s, %d , parentIP: %s, %d \n", a.NodeId, localIP, dmsg.NodeID, parentIP)
	mdis := int64(parentIP) - int64(localIP)
	fmt.Printf("localIP: %s, %d , parentIP: %s, %d， distance: %d \n", a.NodeId, localIP, dmsg.NodeID, parentIP, mdis)
	if mdis < int64(-1*distance) || mdis > int64(distance) {
		return
	}
	fmt.Println(a.NodeId, "handle ", dmsg.NodeID)
	// 加入一跳桶
	a.WriteMsg(dmsg, ep)
	a.Graph.AddEdge(a.NodeId, dmsg.NodeID)

	rev, exist := func() (int, bool) {
		value, ok := a.NodeBuf.Load(dmsg.NodeID)
		if ok {
			return value.(int), true
		}
		return 0, false
	}()
	if exist {
		if rev < dmsg.Revision {
			a.NodeBuf.Store(dmsg.NodeID, dmsg.Revision)
		}
	} else {
		a.NodeBuf.Store(dmsg.NodeID, dmsg.Revision)
	}

	// add msg
	path := Path{dmsg.NodeID}
	a.UpdateMsgs(dmsg, path, ep)

	// handle other msg
	for _, m := range msg.Msgs {
		a.Graph.AddEdge(dmsg.NodeID, m.PrevNode)
		// add PrevAdj
		for _, pn := range m.PrevAdj {
			a.Graph.AddEdge(m.PrevNode, pn)
		}
		// handle msg
		if !common.IsStructEmpty(m.NodeMsg) {
			a.WriteMsg(m.NodeMsg, ep)
			path = Path{m.PrevNode, dmsg.NodeID}
			if m.NodeMsg.NodeID != a.NodeId {
				a.UpdateMsgs(m.NodeMsg, path, ep)
			}
		}
	}
}

// 处理接收到的Gossip消息
func (a *Agent) PathExistInMLST(g *common.Graph, p Path) bool {

	preNode := p[0]
	mlst, _ := g.MLST10(preNode)
	fmt.Println(a.NodeId, " root: ", preNode, " path: ", p, " mlst: ")
	//mlst.Display()
	// if node is leaf, return false
	if mlst.IsLeaf(a.NodeId) {
		return false
	}

	if mlst.PathExists(p) {
		return true
	}
	return false
}

func (a *Agent) UpdateMsgs(msg common.NodeMessage, path Path, ep int) {
	if msg.Revision >= ep {
		return
	}
	if a.isMsgRecorded(msg.NodeID, msg.Revision) {
		return
	}
	key := fmt.Sprintf("%s_%d", msg.NodeID, msg.Revision)
	value, exist := a.Msgs.Load(key)
	Hm := HostMsg{
		Msg: msg,
	}

	if exist {
		existingHostMsg := value.(HostMsg)
		sendpath := append(existingHostMsg.SendPaths, path)
		Hm.SendPaths = sendpath
	} else {
		Hm.SendPaths = []Path{path}
	}

	a.Msgs.Store(key, Hm)
}
