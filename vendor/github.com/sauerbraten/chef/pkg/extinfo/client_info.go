package extinfo

import (
	"errors"
	"fmt"
	"net"

	"github.com/sauerbraten/cubecode"
)

// ClientInfo contains the raw information sent back from the server, i.e. state and privilege are ints.
type ClientInfo struct {
	ClientNum int       `json:"cn"`        //
	Ping      int       `json:"ping"`      // client's ping to server
	Name      string    `json:"name"`      //
	Team      string    `json:"team"`      // name of the team the client is on, e.g. "good"
	Frags     int       `json:"frags"`     // aka kills
	Flags     int       `json:"flags"`     //
	Deaths    int       `json:"deaths"`    //
	Teamkills int       `json:"teamkills"` //
	Accuracy  int       `json:"accuracy"`  // damage the client could have dealt * 100 / damage actually dealt by the client
	Health    int       `json:"health"`    //
	Armour    int       `json:"armour"`    //
	Weapon    Weapon    `json:"weapon"`    //
	Privilege Privilege `json:"privilege"` // 0 ("none"), 1 ("master"), 2 ("auth") or 3 ("admin")
	State     State     `json:"state"`     // client state, e.g. 1 ("alive") or 5 ("spectator"), see names.go for int -> string mapping
	IP        net.IP    `json:"ip"`        // only the first 3 bytes
}

// GetClientInfo returns the raw information about the client with the given CN.
func (s *Server) GetClientInfo(cn int) (map[int]*ClientInfo, error) {
	request := []byte{InfoTypeExtended, ExtInfoTypeClientInfo, byte(cn)}

	c, done, err := s.pinger.send(s.host, s.port, request, s.timeOut)
	if err != nil {
		return nil, err
	}
	defer done()

	resp, ok := <-c
	if !ok {
		return nil, fmt.Errorf("receiving response from %s:%d timed out", s.host, s.port)
	}

	clientNumList, err := parseResponse(request, resp)
	if err != nil {
		return nil, err
	}

	cns, err := parseClientNums(clientNumList)
	if err != nil {
		return nil, err
	}

	stats := map[int]*ClientInfo{}
	for range cns {
		resp, ok = <-c
		if !ok {
			return nil, fmt.Errorf("receiving response from %s:%d timed out", s.host, s.port)
		}

		clientStats, err := parseResponse(request, resp)
		if err != nil {
			return nil, err
		}

		info, err := parseClientStats(clientStats)
		if err != nil {
			return nil, err
		}

		stats[info.ClientNum] = info
	}

	return stats, nil
}

func parseClientNums(response *cubecode.Packet) (cns []int, err error) {
	// expect ClientInfoResponseTypeCNs
	packetType, err := response.ReadInt()
	if err != nil {
		return nil, fmt.Errorf("extinfo: reading client info packet type: %w", err)
	}
	if packetType != ClientInfoResponseTypeCNs {
		return nil, fmt.Errorf("extinfo: parsing client info packet: expected type %d, but got %d", ClientInfoResponseTypeCNs, packetType)
	}

	cns = []int{}
	for response.HasRemaining() {
		cn, err := response.ReadInt()
		if err != nil {
			return nil, fmt.Errorf("extinfo: reading CN from client info packet: %w", err)
		}

		cns = append(cns, cn)
	}

	return cns, nil
}

// own function, because it is used in GetClientInfo() & GetAllClientInfo()
func parseClientStats(response *cubecode.Packet) (clientInfo *ClientInfo, err error) {
	// expect ClientInfoResponseTypeStats
	packetType, err := response.ReadInt()
	if err != nil {
		return nil, errors.New("extinfo: reading client info packet type: " + err.Error())
	}
	if packetType != ClientInfoResponseTypeStats {
		return nil, fmt.Errorf("extinfo: parsing client info packet: expected type %d, but got %d", ClientInfoResponseTypeStats, packetType)
	}

	clientInfo = &ClientInfo{}

	clientInfo.ClientNum, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading client number: " + err.Error())
		return
	}

	clientInfo.Ping, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading ping: " + err.Error())
		return
	}

	clientInfo.Name, err = response.ReadString()
	if err != nil {
		err = errors.New("extinfo: reading client name: " + err.Error())
		return
	}

	clientInfo.Team, err = response.ReadString()
	if err != nil {
		err = errors.New("extinfo: reading team: " + err.Error())
		return
	}

	clientInfo.Frags, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading frags: " + err.Error())
		return
	}

	clientInfo.Flags, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading flags: " + err.Error())
		return
	}

	clientInfo.Deaths, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading deaths: " + err.Error())
		return
	}

	clientInfo.Teamkills, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading teamkills: " + err.Error())
		return
	}

	clientInfo.Accuracy, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading accuracy: " + err.Error())
		return
	}

	clientInfo.Health, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading health: " + err.Error())
		return
	}

	clientInfo.Armour, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading armour: " + err.Error())
		return
	}

	w, err := response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading weapon in use: " + err.Error())
		return
	}
	clientInfo.Weapon = Weapon(w)

	p, err := response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading client privilege: " + err.Error())
		return
	}
	clientInfo.Privilege = Privilege(p)

	s, err := response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading client state: " + err.Error())
		return
	}
	clientInfo.State = State(s)

	// IP from next 3 bytes (sauer never sends 4th IP byte)
	var ipByte1, ipByte2, ipByte3 byte

	ipByte1, err = response.ReadByte()
	if err != nil {
		err = errors.New("extinfo: reading first IP byte: " + err.Error())
		return
	}

	ipByte2, err = response.ReadByte()
	if err != nil {
		err = errors.New("extinfo: reading second IP byte: " + err.Error())
		return
	}

	ipByte3, err = response.ReadByte()
	if err != nil {
		err = errors.New("extinfo: reading third IP byte: " + err.Error())
		return
	}

	clientInfo.IP = net.IPv4(ipByte1, ipByte2, ipByte3, 0)

	return
}
