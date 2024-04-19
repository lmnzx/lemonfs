package p2p

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opt := TCPTransportOps{
		ListenAddr:    ":3000",
		HandshakeFunc: NOPHandshakeFunc,
		Decoder:       DefaultDecoder{},
	}

	tr := NewTCPTransport(opt)
	assert.Equal(t, tr.ListenAddr, opt.ListenAddr)

	// Server
	assert.Nil(t, tr.ListenAndAccept())
}
