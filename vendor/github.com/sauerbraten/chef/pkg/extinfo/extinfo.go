// Package extinfo provides easy access to the state information of a Sauerbraten game server (called 'extinfo' in the Sauerbraten source code).
package extinfo

import (
	"errors"
	"strconv"

	"github.com/sauerbraten/cubecode"
)

// Protocol constants
const (
	// Constants describing the type of information to query for
	InfoTypeExtended = 0
	InfoTypeBasic    = 1

	// Constants used in responses to extended info queries
	ExtInfoACK     = 255
	ExtInfoVersion = 105
	ExtInfoError   = 1

	// Constants describing the type of extended information to query for
	ExtInfoTypeUptime     = 0
	ExtInfoTypeClientInfo = 1
	ExtInfoTypeTeamScores = 2

	// Constants used in responses to client info queries
	ClientInfoResponseTypeCNs   = -10
	ClientInfoResponseTypeStats = -11
)

func parseResponse(request, response []byte) (packet *cubecode.Packet, err error) {
	// response must include the entire request, ExtInfoAck, ExtInfoVersion, and either ExtInfoError or uptime
	if len(response) < len(request)+3 {
		err = errors.New("extinfo: invalid response: too short")
		return
	}

	// make sure the entire request is correctly replayed
	for i := range request {
		if response[i] != request[i] {
			err = errors.New("extinfo: invalid response: response does not match request")
			return
		}
	}

	infoType := response[0]

	// end of basic info response handling
	if infoType == InfoTypeBasic {
		return cubecode.NewPacket(response[len(request):]), nil
	}

	command := response[1]

	// skip any extra bytes of replayed request
	response = response[len(request):]

	// validate ack
	ack := response[0]
	if ack != ExtInfoACK {
		err = errors.New("extinfo: invalid response: expected " + strconv.Itoa(int(ExtInfoACK)) + " (ACK), got " + strconv.Itoa(int(ack)))
		return
	}

	// validate version
	version := response[1]
	// this package only supports protocol version 105
	if version != ExtInfoVersion {
		err = errors.New("extinfo: wrong version: expected " + strconv.Itoa(int(ExtInfoVersion)) + ", got " + strconv.Itoa(int(version)))
		return
	}

	// end of uptime request handling
	if command == ExtInfoTypeUptime {
		return cubecode.NewPacket(response[2:]), nil
	}

	// check for error
	if response[2] == ExtInfoError {
		switch command {
		case ExtInfoTypeClientInfo:
			err = errors.New("extinfo: no client with cn " + strconv.Itoa(int(request[2])))
		case ExtInfoTypeTeamScores:
			err = errors.New("extinfo: server is not running a team mode")
		}
		return
	}

	return cubecode.NewPacket(response[3:]), nil
}
