package p2p

import "net"

// Interface that represents the remote node
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Handles communication between nodes in the network
// Can be TCP, UDP, websockets, ...
type Transport interface {
	Addr() string
	Dail(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
