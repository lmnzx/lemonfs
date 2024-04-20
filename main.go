package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lmnzx/lemonfs/p2p"
)

func OnPeer(p p2p.Peer) error {
	fmt.Println("new peer connected")
	return nil
}

func main() {
	tcpTransportOpts := p2p.TCPTransportOps{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        OnPeer,
	}

	t := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       "3000_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         t,
	}

	s := NewFileServer(fileServerOpts)

	go func() {
		time.Sleep(time.Second * 10)
		s.Stop()
	}()

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
