package transport

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/pmoieni/quic-example/internal/webtransport"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/sync/errgroup"
)

type HTTP3Server struct {
	wtHub *webtransport.Hub
	http  *http3.Server
}

type ServerFlags struct {
	Host string
	Port uint
}

func NewHTTP3Server(flags *ServerFlags, tlsConfig *tls.Config, quicConfig *quic.Config) *HTTP3Server {
	mux := http.NewServeMux()
	server := &HTTP3Server{
		// TODO: set context
		wtHub: webtransport.NewHub(context.Background()),
		http: &http3.Server{
			Handler:         mux,
			Addr:            flags.Host + fmt.Sprintf(":%d", flags.Port),
			TLSConfig:       tlsConfig,
			QUICConfig:      quicConfig,
			EnableDatagrams: true,
		},
	}

	server.setupHandler(mux)

	return server
}

func (s *HTTP3Server) Run(certPath, keyPath string) error {
	sCtx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	eg, egCtx := errgroup.WithContext(sCtx)
	eg.Go(func() error {
		log.Printf("App server starting on %s", s.http.Addr)

		if certPath != "" || keyPath != "" {
			return s.http.ListenAndServeTLS(certPath, keyPath)
		} else {
			return s.http.ListenAndServe()
		}
	})

	eg.Go(func() error {
		<-egCtx.Done()
		// if context.Background is "Done" or the timeout is exceeded, it'll cause an immediate shutdown
		return s.http.Shutdown(context.Background()) // no idea how much timeout is needed
	})

	return eg.Wait()
}

// TODO
func (s *HTTP3Server) Shutdown(ctx context.Context, timeout time.Duration) error {
	return nil
}

func (s *HTTP3Server) setupHandler(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from HTTP/3! You requested %s via %s\n", r.URL.Path, r.Proto)
	})
	mux.Handle("/datagrams", handleDatagrams())
	mux.Handle("/webtransport", s.wtHub.Middleware(handleWebtransport(s.wtHub)))
}

func handleWebtransport(hub *webtransport.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn := w.(http3.Hijacker).Connection()

		if err := hub.AddStream(conn); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func handleDatagrams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn := w.(http3.Hijacker).Connection()

		select {
		case <-conn.ReceivedSettings():
		case <-time.After(10 * time.Second):
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !conn.Settings().EnableDatagrams {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)

		stream := w.(http3.HTTPStreamer).HTTPStream()

		if err := stream.SendDatagram([]byte("http3: deez")); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed to send datagram: %v", err)
			return
		}

		data, err := stream.ReceiveDatagram(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("failed to receive datagram: %v", err)
			return
		}

		fmt.Println("datagram from client: " + string(data))

		stream.Close()
	}
}
