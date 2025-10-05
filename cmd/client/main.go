package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"time"

	"github.com/quic-go/quic-go"
	"golang.org/x/sync/errgroup"
)

const message = "deez"

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(context.Background(), "localhost:1234", tlsConf, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.CloseWithError(0, "")

	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()

	ticker := time.NewTicker(time.Second)

	eg, _ := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		for range ticker.C {
			if _, err := stream.Write([]byte(message)); err != nil {
				return err
			}
		}

		buf := make([]byte, len(message))

		if _, err := io.ReadFull(stream, buf); err != nil {
			return err
		}

		return nil
	})

	eg.Wait()

	/*
		fmt.Printf("Client: Got '%s'\n", buf)
	*/
}
