const fallbackInfo = {
    description: 'loading...',
    game_mode: 0,
    map: 'firstevermap',
    secs_left: 0,
    master_mode: 0,
    num_clients: 0,
    num_slots: 0
}

export const scoreboard = {
    info: {...fallbackInfo},
    teams: new Map(),
    teamless: [],
    spectators: [],
}

export function resetScoreboard (){
    scoreboard.info = {...fallbackInfo}
    scoreboard.teams.clear()
    scoreboard.teamless = []
    scoreboard.spectators = []
}

export const serverlist = {
    servers: []
}