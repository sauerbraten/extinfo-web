package extinfo

import "github.com/sauerbraten/cubecode"

// GetServerMod returns the name of the mod in use at this server.
func (s *Server) GetServerMod() (string, error) {
	request := []byte{InfoTypeExtended, ExtInfoTypeUptime, 0x01}

	c, err := s.pinger.send(s.host, s.port, request, s.timeOut)
	if err != nil {
		return "", err
	}

	response, err := parseResponse(request, <-c)
	if err != nil {
		return "", err
	}

	// read & discard uptime
	_, err = response.ReadInt()
	if err != nil {
		return "", err
	}

	// try to read one more byte
	mod, err := response.ReadInt()

	// if there is none, it's not a detectable mod (probably vanilla), so we will return ""
	if err == cubecode.ErrBufferTooShort {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return getServerModName(mod), nil
}
