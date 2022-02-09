package extinfo

import (
	"fmt"
	"time"

	"github.com/sauerbraten/cubecode"
)

// GetServerMod returns the name of the mod in use at this server.
func (s *Server) GetServerMod() (ServerMod, error) {
	request := []byte{InfoTypeExtended, ExtInfoTypeUptime, 0x01}

	c, done, err := s.pinger.send(s.host, s.port, request, s.timeOut)
	if err != nil {
		return 0, err
	}

	var resp []byte
	select {
	case <-time.After(5 * time.Second):
		return 0, fmt.Errorf("receiving response from %s:%d timed out", s.host, s.port)
	case resp = <-c:
	}
	response, err := parseResponse(request, resp)
	done()
	if err != nil {
		return 0, err
	}

	// read & discard uptime
	_, err = response.ReadInt()
	if err != nil {
		return 0, err
	}

	// try to read one more byte
	mod, err := response.ReadInt()

	// if there is none, it's not a detectable mod (probably vanilla), so we will return ""
	if err == cubecode.ErrBufferTooShort {
		return 0, nil
	} else if err != nil {
		return 0, err
	}

	return ServerMod(mod), nil
}
