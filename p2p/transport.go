package p2p

import "net"

// Interface that represents the remote node
type Peer interface {
	RemoteAddr() net.Addr
	Close() error
}

// Handles communication between nodes in the network
// Can be TCP, UDP, websockets, ...
type Transport interface {
	Dail(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
