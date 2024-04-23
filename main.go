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
	s3 := makeServer(":8000", ":3000", ":4000")

	go s1.Start()
	time.Sleep(2 * time.Second)

	go s2.Start()
	time.Sleep(2 * time.Second)

	go s3.Start()
	time.Sleep(2 * time.Second)

	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("privatedata_%d.txt", i)

		data := bytes.NewReader([]byte("lemon loves you"))
		s3.Store(key, data)

		if err := s3.store.Delete(key); err != nil {
			log.Fatal(err)
		}
		fmt.Println("File deleted from s1")

		r, err := s3.Get(key)
		if err != nil {
			log.Fatal(err)
		}

		b, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("------>", string(b))
		time.Sleep(1 * time.Second)
	}
}
