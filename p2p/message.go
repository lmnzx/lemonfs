package p2p

import "net"

// Represents any arbitrary data that is being
// sent over the transport between the nodes
type Message struct {
	From    net.Addr
	Payload []byte
}
