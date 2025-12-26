package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/sync/errgroup"
)

const message = "deez"

func main() {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{http3.NextProtoH3},
	}

	/*
		stream, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		defer stream.Close()

		ticker := time.NewTicker(time.Second)


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
	*/

	eg, _ := errgroup.WithContext(context.Background())

	http3Transport := &http3.Transport{
		TLSClientConfig: tlsConf,
		EnableDatagrams: true,
	}
	defer http3Transport.Close()

	conn, err := quic.DialAddr(context.Background(), "localhost:1234", tlsConf, &quic.Config{EnableDatagrams: true})
	if err != nil {
		log.Fatalf("failed to dial quic address: %v", err)
	}

	defer conn.CloseWithError(0, "")

	http3Conn := http3Transport.NewClientConn(conn)

	select {
	case <-http3Conn.ReceivedSettings():
	case <-http3Conn.Context().Done():
		log.Println("http3 connection closed")
		return
	}
	settings := http3Conn.Settings()
	if !settings.EnableDatagrams {
		log.Println("no http3 datagrams support")
		return
	}

	http3Stream, err := http3Conn.OpenRequestStream(context.Background())
	if err != nil {
		log.Fatalf("failed to open request stream: %v", err)
	}

	u, _ := url.Parse("https://localhost:1234/datagrams")

	if err := http3Stream.SendRequestHeader(&http.Request{Method: http3.MethodHead0RTT, URL: u}); err != nil {
		log.Fatalf("failed to send request header: %v", err)
	}

	eg.Go(func() error {
		if err := http3Stream.SendDatagram([]byte("http3 client message")); err != nil {
			return fmt.Errorf("unable to send datagram: %v", err)
		}

		data, err := http3Stream.ReceiveDatagram(context.Background())
		if err != nil {
			return fmt.Errorf("unable to receive datagram: %v", err)
		}
		fmt.Println(string(data))

		return nil
	})

	res, err := http3Stream.ReadResponse()
	if err != nil {
		log.Fatalf("failed to read serer response: %v", err)
	}

	fmt.Println("response from server: " + res.Status)

	eg.Wait()

	/*
		fmt.Printf("Client: Got '%s'\n", buf)
	*/
}
