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
		model.info = {
			description: 'loading …',
			gameMode: 'ffa',
			map: 'firstevermap',
			secsLeft: 0,
			masterMode: 'open',
			numberOfClients: 0,
			maxNumberOfClients: 0
		}
		model.teams = {}
		model.spectators = []
	}

	addr = window.location.hash.substring(1)
	initsocket()

	window.scrollTo(window.scrollLeft, 0)
}

function initsocket() {
	sock = new WebSocket(protocol + '//' + window.location.host + '/server/' + addr)

	sock.onerror = () => alert('could not connect to a server at that address!')

	sock.onmessage = m => {
		var update = JSON.parse(m.data)

		model.info = update.serverinfo

		document.title = update.serverinfo.description + ' – extinfo'
		spectators = []

		for (teamName in update.teams) {
			update.teams[teamName].players = []
		}

		for (cn in update.players) {
			var player = update.players[cn]
			if (player.state == 'spectator') {
				spectators.push(player)
			} else {
				update.teams[player.team].players.push(player)
			}
		}

		model.spectators = spectators

		for (teamName in update.teams) {
			update.teams[teamName].players.sort(scoreboardSortingFunction)
		}

		model.teams = update.teams
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
	master = new WebSocket(protocol + '//' + window.location.host + '/master')

	master.onerror = () => alert('could not connect to the master server!')

	master.onmessage = m => {
		var serverlistUpdate = JSON.parse(m.data)
		serverlistUpdate.sort((a, b) => b.numberOfClients - a.numberOfClients)
		model.servers = serverlistUpdate
	}
}
