package extinfo

import (
	"errors"
	"time"
)

// Server represents a Sauerbraten game server.
type Server struct {
	pinger  *Pinger
	host    string
	port    int
	timeOut time.Duration
}

// NewServer returns a Server to query information from.
func NewServer(p *Pinger, host string, port int, timeOut time.Duration) (*Server, error) {
	if p == nil {
		return nil, errors.New("nil Pinger provided to NewServer")
	}

	return &Server{
		pinger:  p,
		host:    host,
		port:    port,
		timeOut: timeOut,
	}, nil
}
