package gossip

// Gossip消息
type NodeMessage struct {
	NodeID   string
	Revision int
	Data     map[string]string
}

type GossipMessage struct {
	Self   NodeMessage
	Direct []NodeMessage
	Other  []NodeMessage
}
