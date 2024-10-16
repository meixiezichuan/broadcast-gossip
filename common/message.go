package common

// Gossip消息
type NodeMessage struct {
	NodeID   string
	Revision int
	Data     map[string]string
}

type GossipMessage struct {
	Self NodeMessage
	Msgs []SendMessage
}

type SendMessage struct {
	PrevNode string
	NodeMsg  NodeMessage
}
