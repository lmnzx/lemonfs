package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
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
	net.Conn
	// If we dial and retrieve a conn => outbound == true
	// If we accept and retrieve a conn => outbound == false
	outbound bool

	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}

func (p *TCPPeer) CloseStream() {
	p.wg.Done()
}

func NewTCPTransport(opts TCPTransportOps) *TCPTransport {
	return &TCPTransport{
		TCPTransportOps: opts,
		rpcch:           make(chan RPC, 1024),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("tcp transport listening on port: %s\n", t.ListenAddr)

	return nil
}

func (t *TCPTransport) Addr() string {
	return t.ListenAddr
}

// Dail implements the Transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

// Consume implements the Transport interface
// retrun read-only channel for reading the incoming
// messages received from other peers in the network
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcch
}

// Close closes the tcp listener
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("tcp accept error: %s\n", err)
		}

		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) error {
	var err error

	defer func() {
		fmt.Printf("dropping peer connection: %s\n", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return err
	}

	if t.OnPeer != nil {
		if err := t.OnPeer(peer); err != nil {
			return err
		}
	}

	// Read loop
	for {
		rpc := RPC{}

		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("tcp error: %s\n", err)
			return nil
		}

		rpc.From = conn.RemoteAddr()

		if rpc.Stream {

			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())

			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		t.rpcch <- rpc
	}
}
