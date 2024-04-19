package p2p

import (
	"fmt"
	"net"
)

// Configuration for a TCP transport
type TCPTransportOps struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOps
	listener net.Listener
	rpcch    chan RPC
}

// Represents the remote node over a TCP established connection
type TCPPeer struct {
	// Underlying connection of the peer
	conn net.Conn
	// If we dial and retrieve a conn => outbound == true
	// If we accept and retrieve a conn => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

func NewTCPTransport(opts TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOps: opts,
		rpcch:           make(chan RPC),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

// Implements the Transport interface
// retrun read-only channel for reading the incoming
// messages received from other peers in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		fmt.Printf("New incoming connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) error {
	var err error

	defer func() {
		fmt.Printf("Dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err = t.HandshakeFunc(peer); err != nil {
		return err
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return err
		}
	}

	// Read loop
	rpc := RPC{}
	for {
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("TCP error: %s\n", err)
			return nil
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
