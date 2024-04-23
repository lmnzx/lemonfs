package p2p

import "net"

// Interface that represents the remote node
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Handles communication between nodes in the network
// Can be TCP, UDP, websockets, ...
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
