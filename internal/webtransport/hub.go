package webtransport

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/quic-go/quic-go/http3"
)

type clientIDKey struct{}

// clientTracingID should be derived from HTTP3 connection context
type clientTracingID uuid.UUID

// hub manages all the sessions
type Hub struct {
	ctx      context.Context
	mux      sync.Mutex
	sessions map[clientTracingID]map[sessionID]*Session
}

func NewHub(ctx context.Context) *Hub {
	return &Hub{
		ctx:      ctx,
		sessions: make(map[clientTracingID]map[sessionID]*Session),
	}
}

func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracingID, err := uuid.NewUUID()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.WithValue(r.Context(), clientIDKey{}, tracingID)

		r = r.WithContext(ctx)

	}
}

func (h *Hub) AddStream(conn *http3.Conn) error {

}

func (h *Hub) RemoveSession() {

}
