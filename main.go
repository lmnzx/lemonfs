package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/lmnzx/lemonfs/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOps{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}

	t := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       listenAddr + "_lemonfs",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         t,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	t.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "")
	s2 := makeServer(":4000", ":3000")

	// go func() { log.Fatal(s1.Start()) }()
	go s1.Start()

	go s2.Start()

	time.Sleep(1 * time.Second)

	key := "privatedata_ttt"

	data := bytes.NewReader([]byte("big data askdjflkasjdfl askdfjlaksjdfklajsdflkajdf"))
	s1.Store(key, data)

	if err := s1.store.Delete(key); err != nil {
		log.Fatal(err)
	}
	fmt.Println("File deleted from s1")

	time.Sleep(10 * time.Second)

	r, err := s1.Get(key)
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("------>", string(b))
	select {}
}
