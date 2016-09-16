var sock
var port = '28785'
var host = 'pastaland.ovh'

function init() {
	if (!('WebSocket' in window)) {
		error('sorry, but your browser does not support websockets', 'try updating your browser')
		return
	}

	if (window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
		var parts = window.location.hash.substring(1).split(':')
		host = parts[0]
		port = parts[1]
	}

	initsocket()
}

function initsocket() {
	if (typeof (sock) != 'undefined') {
		sock.close()
		sock = null
	}

	sock = new WebSocket('ws://' + window.location.host + '/ws')

	sock.onopen = function (e) {
		console.log(' - socket opened - ')
		sock.send(host + ':' + port)
		console.log('    sent:', host + ':' + port)
	}

	sock.onclose = function (e) {
		console.log(' - socket closed - ')
		document.title = 'extinfo-web'
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

function error(err) {
	console.log(err)
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