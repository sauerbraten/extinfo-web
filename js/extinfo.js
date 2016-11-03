var sock
var addr = 'pastaland.ovh:28785'

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
	console.log("resetting")

	if (typeof (sock) != 'undefined') {
		sock.close()
		sock = null

		document.title = 'loading… – extinfo'
		model.info = {
			description: 'loading…',
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

	if (window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
		addr = window.location.hash.substring(1)
		initsocket()
	}
}

function initsocket() {
	sock = new WebSocket('ws://' + window.location.host + '/server/' + addr)

	sock.onerror = function () {
		alert('could not connect to a server at that address!')
	}

	sock.onmessage = function (m) {
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

function initmaster() {
	master = new WebSocket('ws://' + window.location.host + '/master')

	master.onerror = function () {
		alert('could not connect to the master server!')
	}

	master.onmessage = function (m) {
		var serverlistUpdate = JSON.parse(m.data)
		serverlistUpdate.sort(serverlistSortingFunction)
		model.servers = serverlistUpdate
	}
}

// frags (descending), then deaths (ascending), then accuracy (descending)
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

// players (descending)
function serverlistSortingFunction(a, b) {
	return b.numberOfClients - a.numberOfClients
}