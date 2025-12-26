package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"os/signal"
	"syscall"

	"github.com/pmoieni/quic-example/internal/transport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/sync/errgroup"
)

func main() {
	sCtx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	eg, _ := errgroup.WithContext(sCtx)

	tlsConf := generateTLSConfig()

	/*
		srv := transport.NewQuicServer(&transport.QuicServerFlags{Host: "localhost", Port: 1234})

		eg.Go(func() error {
			return srv.Listen(sCtx, tlsConf, &quic.Config{Allow0RTT: true})
		})
	*/

	http3Srv := transport.NewHTTP3Server(&transport.ServerFlags{Host: "localhost", Port: 1234}, tlsConf, &quic.Config{Allow0RTT: true, EnableDatagrams: true})

	eg.Go(func() error {
		return http3Srv.Run("", "")
	})

	eg.Wait()
}

func generateTLSConfig() *tls.Config {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		panic(err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{{
			Certificate: [][]byte{certDER},
			PrivateKey:  priv,
		}},
		NextProtos: []string{http3.NextProtoH3},
	}
}
