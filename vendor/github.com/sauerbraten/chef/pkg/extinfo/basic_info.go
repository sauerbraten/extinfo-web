package extinfo

import (
	"errors"

	"github.com/sauerbraten/cubecode"
)

// BasicInfo contains the information sent by the server in response to a basic info request.
type BasicInfo struct {
	NumberOfClients int        `json:"num_clients"`      // the number of clients currently connected to the server (players and spectators)
	ProtocolVersion int        `json:"protocol_version"` // version number of the protocol in use by the server
	GameMode        GameMode   `json:"game_mode"`        // current game mode
	SecsLeft        int        `json:"secs_left"`        // the time left until intermission in seconds
	NumberOfSlots   int        `json:"num_slots"`        // the maximum number of clients the server allows
	MasterMode      MasterMode `json:"master_mode"`      // the current master mode of the server
	Paused          bool       `json:"paused"`           // wether the game is paused or not
	GameSpeed       int        `json:"game_speed"`       // the gamespeed
	Map             string     `json:"map"`              // current map
	Description     string     `json:"description"`      // server description
}

// GetBasicInfo queries a Sauerbraten server at addr on port and returns the response or an error in case something went wrong.
func (s *Server) GetBasicInfo() (basicInfo *BasicInfo, err error) {
	request := []byte{InfoTypeBasic}

	resp, err := s.pinger.expectSinglePacket(s.host, s.port, request, s.timeOut)
	if err != nil {
		return nil, err
	}

	response, err := parseResponse(request, resp)
	if err != nil {
		return nil, err
	}

	basicInfo = &BasicInfo{}

	basicInfo.NumberOfClients, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading number of connected clients: " + err.Error())
		return
	}

	// next int is always 5 or 7, the number of additional attributes after the clientcount and before the strings for map and description
	sevenAttributes := false
	numberOfAttributes, err := response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading number of following values: " + err.Error())
		return
	}

	if numberOfAttributes == 7 {
		sevenAttributes = true
	}

	basicInfo.ProtocolVersion, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading protocol version: " + err.Error())
		return
	}

	gm, err := response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading game mode: " + err.Error())
		return
	}
	basicInfo.GameMode = GameMode(gm)

	basicInfo.SecsLeft, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading time left: " + err.Error())
		return
	}

	basicInfo.NumberOfSlots, err = response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading maximum number of clients: " + err.Error())
		return
	}

	mm, err := response.ReadInt()
	if err != nil {
		err = errors.New("extinfo: reading master mode: " + err.Error())
		return
	}
	basicInfo.MasterMode = MasterMode(mm)

	if sevenAttributes {
		var isPausedValue int
		isPausedValue, err = response.ReadInt()
		if err != nil {
			err = errors.New("extinfo: reading paused value: " + err.Error())
			return
		}

		if isPausedValue == 1 {
			basicInfo.Paused = true
		}

		basicInfo.GameSpeed, err = response.ReadInt()
		if err != nil {
			err = errors.New("extinfo: reading game speed: " + err.Error())
			return
		}
	} else {
		basicInfo.GameSpeed = 100
	}

	mapname, err := response.ReadString()
	if err != nil {
		err = errors.New("extinfo: reading map name: " + err.Error())
		return
	}
	basicInfo.Map = cubecode.SanitizeString(mapname)

	description, err := response.ReadString()
	if err != nil {
		err = errors.New("extinfo: reading server description: " + err.Error())
		return
	}
	basicInfo.Description = cubecode.SanitizeString(description)

	return
}
