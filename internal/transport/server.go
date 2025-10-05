package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/quic-go/quic-go"
)

const (
	requestUpperBound = 256
)

type QuicServer struct{}

func NewQuicServer() *QuicServer {
	return &QuicServer{}
}

func (s *QuicServer) Listen(ctx context.Context, addr string, tlsConf *tls.Config) error {
	ln, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		return err
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept(ctx)
		if err != nil {
			log.Fatalf("listener failed to accept: %v", err) // TODO: pass error to channel
		}

		go handleStream(conn)
	}
}

func handleStream(conn *quic.Conn) {
	for {
		stream, err := conn.AcceptStream(conn.Context())
		if err != nil {
			log.Fatalf("connection failed to accept stream: %v", err)
		}

		go handleRequest(stream)
	}
}

func handleRequest(stream *quic.Stream) {
	for {
		bs := make([]byte, 256)
		if _, err := stream.Read(bs); err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(bs))
	}
}
