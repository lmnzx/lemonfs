package p2p

import "net"

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
)

// Represents any arbitrary data that is being
// sent over the transport between the nodes
type RPC struct {
	From    net.Addr
	Payload []byte
	Stream  bool
}
