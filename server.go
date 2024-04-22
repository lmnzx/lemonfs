package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/lmnzx/lemonfs/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}
	return &FileServer{
		store:          NewStore(storeOpts),
		FileServerOpts: opts,
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

// Store the file to the disk and broadcast to the peers
func (s *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	size, err := s.store.Write(key, tee)
	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: size,
		},
	}

	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(msgBuf.Bytes()); err != nil {
			return err
		}
	}

	time.Sleep(1 * time.Second)

	for _, peer := range s.peers {
		n, err := io.Copy(peer, buf)
		if err != nil {
			return err
		}

		fmt.Printf("recv and written %d bytes to the disk\n", n)
	}

	return nil
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Println(err)
			}

			if err := s.handleMessage(rpc.From.String(), &msg); err != nil {
				log.Println(err)
			}

		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(form string, msg *Message) error {
	switch v := msg.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(form, v)
	}
	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer (%s) could not be found in the peer list", from)
	}

	if _, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size)); err != nil {
		return nil
	}

	peer.(*p2p.TCPPeer).Wg.Done()

	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addr string) {
			fmt.Printf("[%s] attemping to connect with remote %s\n", s.Transport.Addr(), addr)
			if err := s.Transport.Dail(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}
	return nil
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
}
