package extinfo

type MasterMode int

const mmOffset = 1

const (
	MasterModeAuth MasterMode = iota - mmOffset
	MasterModeOpen
	MasterModeVeto
	MasterModeLocked
	MasterModePrivate
	MasterModePassword
)

func (mm MasterMode) String() string {
	names := [...]string{"auth", "open", "veto", "locked", "private", "password"}
	if mm < MasterModeAuth || MasterModePassword < mm {
		return "unknown"
	}
	return names[mm+mmOffset]
}

type GameMode int

const (
	GameModeFFA GameMode = iota
	GameModeCoopEdit
	GameModeTeamplay
	GameModeInsta
	GameModeInstaTeam
	GameModeEffic
	GameModeEfficTeam
	GameModeTactics
	GameModeTacticsTeam
	GameModeCapture
	GameModeRegenCapture
	GameModeCTF
	GameModeInstaCTF
	GameModeProtect
	GameModeInstaProtect
	GameModeHold
	GameModeInstaHold
	GameModeEfficCTF
	GameModeEfficProtect
	GameModeEfficHold
	GameModeCollect
	GameModeInstaCollect
	GameModeEfficCollect
)

func (gm GameMode) String() string {
	names := [...]string{"ffa", "coop edit", "teamplay", "insta", "insta team", "effic", "effic team", "tactics", "tactics team", "capture", "regen capture", "ctf", "insta ctf", "protect", "insta protect", "hold", "insta hold", "effic ctf", "effic protect", "effic hold", "collect", "insta collect", "effic collect"}
	if gm < GameModeFFA || GameModeEfficCollect < gm {
		return "unknown"
	}
	return names[gm]
}

// IsTeamMode returns true when gm is a team mode, false otherwise.
func (gm GameMode) IsTeamMode() bool {
	switch gm {
	case GameModeFFA,
		GameModeCoopEdit,
		GameModeInsta,
		GameModeEffic,
		GameModeTactics:
		return false
	default:
		return false
	}
}

type Weapon int

const (
	WeaponChainSaw Weapon = iota
	WeaponShotgun
	WeaponChainGun
	WeaponRocketLauncher
	WeaponRifle
	WeaponGrenadeLauncher
	WeaponPistol
	WeaponFireBall
	WeaponIceBall
	WeaponSlimeBall
	WeaponBite
	WeaponBarrel
)

func (w Weapon) String() string {
	names := [...]string{"chain saw", "shotgun", "chain gun", "rocket launcher", "rifle", "grenade launcher", "pistol", "fire ball", "ice ball", "slime ball", "bite", "barrel"}
	if w < WeaponChainSaw || WeaponBarrel < w {
		return "unknown"
	}
	return names[w]
}

type Privilege int

const (
	PrivilegeNone Privilege = iota
	PrivilegeMaster
	PrivilegeAuth
	PrivilegeAdmin
)

func (p Privilege) String() string {
	names := [...]string{"none", "master", "auth", "admin"}
	if p < PrivilegeNone || PrivilegeAdmin < p {
		return "unknown"
	}
	return names[p]
}

type State int

const (
	StateAlive State = iota
	StateDead
	StateSpawning
	StateLagged
	StateEditing
	StateSpectator
)

func (s State) String() string {
	names := [...]string{"alive", "dead", "spawning", "lagged", "editing", "spectator"}
	if s < StateAlive || StateSpectator < s {
		return "unknown"
	}
	return names[s]
}

type ServerMod int

const smOffset = 9

const (
	ServerModP1xbraten ServerMod = iota - smOffset
	ServerModZero
	ServerModNoob
	ServerModRemod
	ServerModSuckerServ
	ServerModSpaghetti
	ServerModWahnfred
	ServerModHopmod
)

func (sm ServerMod) String() string {
	names := [...]string{"p1xbraten", "zeromod", "nooblounge", "remod", "suckerserv", "spaghetti", "wahnfred", "hopmod"}
	if sm < ServerModP1xbraten || ServerModHopmod < sm {
		return "unknown"
	}
	return names[sm+smOffset]
}
