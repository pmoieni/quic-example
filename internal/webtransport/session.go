package webtransport

import "github.com/google/uuid"

type sessionID uuid.UUID

// session manages all the streams and defines a session
type Session struct {
}

func craeteSession() *Session {
	return &Session{}
}

func (s *Session) Terminate() error {

}

func (s *Session) Drain() error {

}

// TODO: no idea what this is used for
func (s *Session) ExportKeyingMaterial() error { return nil }
