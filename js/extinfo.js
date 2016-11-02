var sock
var port = '28785'
var host = 'pastaland.ovh'

function init() {
	if (!('WebSocket' in window)) {
		alert('sorry, but your browser does not support websockets', 'try updating your browser')
		return
	}

	if (window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
		var parts = window.location.hash.substring(1).split(':')
		host = parts[0]
		port = parts[1]
	} else {
		window.location.hash = host + ":" + port
	}

	initsocket()
}

function initsocket() {
	if (typeof (sock) != 'undefined') {
		sock.close()
		sock = null
	}

	sock = new WebSocket('ws://' + window.location.host + '/server/'+host+":"+port)

	sock.onerror = function() {
		alert("could not connect to a server at that address!")
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

// frags (descending), then deaths (ascending), then accuracy (descending)
function scoreboardSortingFunction (a, b) {
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