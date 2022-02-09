package extinfo

import (
	"fmt"
	"time"
)

// TeamScore contains the name of the team and the score, i.e. flags scored in flag modes / points gained for holding bases in capture modes / frags achieved in DM modes / skulls collected.
type TeamScore struct {
	Name  string `json:"name"`  // name of the team, e.g. "good"
	Score int    `json:"score"` // flags in ctf modes, frags in deathmatch modes, points in capture, skulls in collect
	Bases []int  `json:"bases"` // the numbers/IDs of the bases the team possesses (only used in capture modes)
}

// TeamScores contains the game mode, the seconds left in the game, and a slice of TeamScores.
type TeamScores struct {
	GameMode GameMode             `json:"game_mode"` // current game mode
	SecsLeft int                  `json:"secs_left"` // the time left until intermission in seconds
	Scores   map[string]TeamScore `json:"scores"`    // a team score for each team, mapped to the team's name
}

// GetTeamScores queries a Sauerbraten server at addr on port for the teams' names and scores and returns the response and/or an error in case something went wrong or the server is not running a team mode.
func (s *Server) GetTeamScores() (teamScores *TeamScores, err error) {
	request := []byte{InfoTypeExtended, ExtInfoTypeTeamScores}

	c, done, err := s.pinger.send(s.host, s.port, request, s.timeOut)
	if err != nil {
		return nil, err
	}

	var resp []byte
	select {
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("receiving response from %s:%d timed out", s.host, s.port)
	case resp = <-c:
	}
	response, err := parseResponse(request, resp)
	done()
	if err != nil {
		return nil, err
	}

	teamScores = &TeamScores{}

	gm, err := response.ReadInt()
	if err != nil {
		return
	}
	teamScores.GameMode = GameMode(gm)

	teamScores.SecsLeft, err = response.ReadInt()
	if err != nil {
		return
	}

	teamScores.Scores = map[string]TeamScore{}

	for response.HasRemaining() {
		var name string
		name, err = response.ReadString()
		if err != nil {
			return
		}

		var score int
		score, err = response.ReadInt()
		if err != nil {
			return
		}

		var numBases int
		numBases, err = response.ReadInt()
		if err != nil {
			return
		}

		if numBases < 0 {
			numBases = 0
		}

		bases := make([]int, numBases)

		for i := 0; i < numBases; i++ {
			var base int
			base, err = response.ReadInt()
			if err != nil {
				return
			}
			bases = append(bases, base)
		}

		teamScores.Scores[name] = TeamScore{name, score, bases}
	}

	return
}
