export const scoreboard = {
    info: {
        description: 'loading...',
        gameMode: 'ffa',
        map: 'firstevermap',
        secsLeft: 0,
        masterMode: 'open',
        numberOfClients: 0,
        maxNumberOfClients: 0
    },
    teams: new Map(),
    teamless: [],
    spectators: [],
}

export function resetScoreboard (){
    scoreboard.info = {
        description: 'loading...',
        gameMode: 'ffa',
        map: 'firstevermap',
        secsLeft: 0,
        masterMode: 'open',
        numberOfClients: 0,
        maxNumberOfClients: 0
    }
    scoreboard.teams.clear()
    scoreboard.teamless = []
    scoreboard.spectators = []
}

export const serverlist = {
    servers: []
}