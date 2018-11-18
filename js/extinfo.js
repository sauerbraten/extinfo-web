var sock
var addr = 'pastaland.ovh:28785'
const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'

function init() {
	if (!('WebSocket' in window)) {
		alert('sorry, but your browser does not support websockets', 'try updating your browser')
		return
	}

	initmaster()

	window.onhashchange = reset

	if (window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
		addr = window.location.hash.substring(1)
		initsocket()
	} else {
		window.location.hash = addr
	}
}

function reset() {
	if (typeof sock != 'undefined') {
		sock.close()
		sock.onerror = null
		sock = null

		document.title = 'loading … – extinfo'
		scoreboard.info = {
			description: 'loading …',
			gameMode: 'ffa',
			map: 'firstevermap',
			secsLeft: 0,
			masterMode: 'open',
			numberOfClients: 0,
			maxNumberOfClients: 0
		}
		scoreboard.teams = {}
		scoreboard.teamless = []
		scoreboard.spectators = []
	}

	addr = window.location.hash.substring(1)
	initsocket()

	window.scrollTo(window.scrollLeft, 0)
}

function initsocket() {
	sock = new WebSocket(`${protocol}//${window.location.host}/server/${addr}`)
	window.addEventListener('beforeunload', () => sock.close())

	sock.onerror = () => alert('could not connect to a server at that address!')

	sock.onmessage = m => {
		let update = JSON.parse(m.data)

		scoreboard.info = update.serverinfo
		document.title = update.serverinfo.description + ' – extinfo'

		let teams = {}, teamless = [], spectators = []

		if ('teams' in update) {
			for (const teamName in update.teams) {
				teams[teamName] = update.teams[teamName]
				teams[teamName].players = []
			}
		}

		for (const cn in update.players) {
			let player = update.players[cn]
			if (player.state == 'spectator') {
				spectators.push(player)
			} else if (player.team in teams) {
				teams[player.team].players.push(player)
			} else {
				teamless.push(player)
			}
		}

		if (Object.keys(teams).length > 0) {
			for (const teamName in teams) {
				teams[teamName].players.sort(scoreboardSortingFunction)
			}
			scoreboard.teamless = []
			scoreboard.teams = teams
		} else {
			teamless.sort(scoreboardSortingFunction)
			scoreboard.teams = []
			scoreboard.teamless = teamless
		}
		scoreboard.spectators = spectators
	}
}

// sorts by frags (descending), then deaths (ascending), then accuracy (descending)
function scoreboardSortingFunction(a, b) {
	if (a.frags == b.frags) {
		if (a.deaths == b.deaths) {
			return b.accuracy - a.accuracy
		} else {
			return a.deaths - b.deaths
		}
	} else {
		return b.frags - a.frags
	}
}

function initmaster() {
	master = new WebSocket(`${protocol}//${window.location.host}/master`)
	window.addEventListener('beforeunload', () => master.close())

	master.onerror = () => alert('could not connect to the master server!')

	master.onmessage = m => {
		let update = JSON.parse(m.data)
		update.sort((a, b) => b.numberOfClients - a.numberOfClients)
		serverlist.servers = update
	}
}
