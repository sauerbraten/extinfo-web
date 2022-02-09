const mmOffset = 1  // because -1 = auth
const modOffset = 9 // because -9 = p1xbraten

export const names = {
    mm:    i => ["auth", "open", "veto", "locked", "private", "password"][i+mmOffset],
    gm:    i => ["ffa", "coop edit", "teamplay", "insta", "insta team", "effic", "effic team", "tactics", "tactics team", "capture", "regen capture", "ctf", "insta ctf", "protect", "insta protect", "hold", "insta hold", "effic ctf", "effic protect", "effic hold", "collect", "insta collect", "effic collect"][i],
    priv:  i => ["none", "master", "auth", "admin"][i],
    state: i => ["alive", "dead", "spawning", "lagged", "editing", "spectator"][i],
    mod:   i => ["p1xbraten", "zeromod", "nooblounge", "remod", "suckerserv", "spaghetti", "wahnfred", "hopmod"][i+modOffset],
}