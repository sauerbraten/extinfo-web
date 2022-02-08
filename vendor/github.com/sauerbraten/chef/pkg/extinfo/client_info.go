package extinfo

import (
	"errors"
	"fmt"
	"net"

	"github.com/sauerbraten/cubecode"
)

// ClientInfoRaw contains the raw information sent back from the server, i.e. state and privilege are ints.
type ClientInfoRaw struct {
	ClientNum int    `json:"clientNum"` // client number or cn
	Ping      int    `json:"ping"`      // client's ping to server
	Name      string `json:"name"`      //
	Team      string `json:"team"`      // name of the team the client is on, e.g. "good"
	Frags     int    `json:"frags"`     // kills
	Flags     int    `json:"flags"`     // number of flags the player scored
	Deaths    int    `json:"deaths"`    //
	Teamkills int    `json:"teamkills"` //
	Accuracy  int    `json:"accuracy"`  // damage the client could have dealt * 100 / damage actually dealt by the client
	Health    int    `json:"health"`    // remaining HP (health points)
	Armour    int    `json:"armour"`    // remaining armour
	Weapon    int    `json:"weapon"`    // weapon the client currently has selected
	Privilege int    `json:"privilege"` // 0 ("none"), 1 ("master"), 2 ("auth") or 3 ("admin")
	State     int    `json:"state"`     // client state, e.g. 1 ("alive") or 5 ("spectator"), see names.go for int -> string mapping
	IP        net.IP `json:"ip"`        // client IP (only the first 3 bytes)
}

// ClientInfo contains the parsed information sent back from the server, i.e. weapon, state and privilege are translated into human readable strings.
type ClientInfo struct {
	*ClientInfoRaw
	Weapon    string `json:"weapon"`    // weapon the client currently has selected
	Privilege string `json:"privilege"` // "none", "master" or "admin"
	State     string `json:"state"`     // client state, e.g. "dead" or "spectator"
}

// GetClientInfoRaw returns the raw information about the client with the given CN.
func (s *Server) GetClientInfoRaw(cn int) (map[int]*ClientInfoRaw, error) {
	request := []byte{InfoTypeExtended, ExtInfoTypeClientInfo, byte(cn)}

	c, err := s.pinger.send(s.host, s.port, request, s.timeOut)
	if err != nil {
		return nil, err
	}

	clientNumList, err := parseResponse(request, <-c)
	if err != nil {
		return nil, err
	}

	cns, err := parseClientNums(clientNumList)
	if err != nil {
		return nil, err
	}

	stats := map[int]*ClientInfoRaw{}
	for range cns {
		clientStats, err := parseResponse(request, <-c)
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

// GetClientInfo returns the ClientInfo of all Players (including spectators) as a []ClientInfo
func (s *Server) GetClientInfo(cn int) (map[int]*ClientInfo, error) {
	rawStats, err := s.GetClientInfoRaw(cn)
	if err != nil {
		return nil, err
	}

	stats := map[int]*ClientInfo{}
	for cn, info := range rawStats {
		stats[cn] = &ClientInfo{
			ClientInfoRaw: info,
			Weapon:        getWeaponName(info.Weapon),
			Privilege:     getPrivilegeName(info.Privilege),
			State:         getStateName(info.State),
		}
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
func parseClientStats(response *cubecode.Packet) (clientInfoRaw *ClientInfoRaw, err error) {
	// expect ClientInfoResponseTypeStats
	packetType, err := response.ReadInt()
	if err != nil {
		return nil, errors.New("extinfo: error reading client info packet type: " + err.Error())
	}
	if packetType != ClientInfoResponseTypeStats {
		return nil, fmt.Errorf("extinfo: error parsing client info packet: expected type %d, but got %d", ClientInfoResponseTypeStats, packetType)
	}

	clientInfoRaw = &ClientInfoRaw{}

	clientInfoRaw.ClientNum, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading client number: " + err.Error())
		return
	}

	clientInfoRaw.Ping, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading ping: " + err.Error())
		return
	}

	clientInfoRaw.Name, err = response.ReadString()
	if err != nil {
		err = errors.New("extinfo: error reading client name: " + err.Error())
		return
	}

	clientInfoRaw.Team, err = response.ReadString()
	if err != nil {
		err = errors.New("extinfo: error reading team: " + err.Error())
		return
	}

	clientInfoRaw.Frags, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading frags: " + err.Error())
		return
	}

	clientInfoRaw.Flags, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading flags: " + err.Error())
		return
	}

	clientInfoRaw.Deaths, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading deaths: " + err.Error())
		return
	}

	clientInfoRaw.Teamkills, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading teamkills: " + err.Error())
		return
	}

	clientInfoRaw.Accuracy, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading accuracy: " + err.Error())
		return
	}

	clientInfoRaw.Health, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading health: " + err.Error())
		return
	}

	clientInfoRaw.Armour, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading armour: " + err.Error())
		return
	}

	clientInfoRaw.Weapon, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading weapon in use: " + err.Error())
		return
	}

	clientInfoRaw.Privilege, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading client privilege: " + err.Error())
		return
	}

	clientInfoRaw.State, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: error reading client state: " + err.Error())
		return
	}

	// IP from next 4 bytes
	var ipByte1, ipByte2, ipByte3, ipByte4 byte

	ipByte1, err = response.ReadByte()
	if err != nil {
		err = errors.New("extinfo: error reading first IP byte: " + err.Error())
		return
	}

	ipByte2, err = response.ReadByte()
	if err != nil {
		err = errors.New("extinfo: error reading second IP byte: " + err.Error())
		return
	}

	ipByte3, err = response.ReadByte()
	if err != nil {
		err = errors.New("extinfo: error reading third IP byte: " + err.Error())
		return
	}

	ipByte4 = 0 // sauer never sends 4th IP byte for privacy reasons

	clientInfoRaw.IP = net.IPv4(ipByte1, ipByte2, ipByte3, ipByte4)

	return
}
