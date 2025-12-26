package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"github.com/quic-go/quic-go"
)

const (
	requestUpperBound = 256
)

type QuicServer struct {
	transport *quic.Transport
}

type QuicServerFlags struct {
	Host string
	Port uint
}

func NewQuicServer(flags *QuicServerFlags) *QuicServer {
	resolved, err := net.ResolveUDPAddr("udp4", fmt.Sprintf(flags.Host+":%d", flags.Port))
	if err != nil {
		log.Fatal("failed to resolve address for QUIC listener: %v", err)
	}

	udpConn, err := net.ListenUDP("udp4", resolved)

	transport := quic.Transport{Conn: udpConn}

	return &QuicServer{transport: &transport}
}

func (s *QuicServer) Listen(ctx context.Context, tlsConf *tls.Config, quicConf *quic.Config) error {
	ln, err := s.transport.Listen(tlsConf, quicConf)
	if err != nil {
		log.Fatal("failed to initialize QUIC listener: %v", err)
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
