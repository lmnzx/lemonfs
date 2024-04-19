package p2p

// Interface that represents the remote node
type Peer interface{}

// Handles communication between nodes in the network
// Can be TCP, UDP, websockets, ...
type Transport interface {
	ListenAndAccept() error
}
