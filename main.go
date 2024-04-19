package main

import (
	"fmt"
	"log"

	"github.com/lmnzx/lemonfs/p2p"
)

func OnPeer(p p2p.Peer) error {
	fmt.Println("We Cool")
	p.Close()
	return nil
}

func main() {
	tcpOpts := p2p.TCPTransportOps{
		ListenAddr:    ":3000",
		Decoder:       p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
		OnPeer:        OnPeer,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
