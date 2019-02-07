import { scoreboard, serverlist, resetScoreboard } from './model.js'
import { initSocket, initMasterSocket, free } from './sockets.js'
import './ui.js'

function processServerUpdate(update) {
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

	if (Object.keys(teams).length) {
		scoreboard.teamless = []
		scoreboard.teams = teams
	} else {
		scoreboard.teams = []
		scoreboard.teamless = teamless
	}
	scoreboard.spectators = spectators
}

function init() {
	if (!('WebSocket' in window)) {
		alert('sorry, but your browser does not support websockets', 'try updating your browser')
		return
	}

	let sock = null
	const socketFromHash = () => {
		if (!window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
			return
		}
		free(sock)
		resetScoreboard()
		document.title = 'loading … – extinfo'
		window.scrollTo(window.scrollLeft, 0)
		sock = initSocket(window.location.hash.substring(1), processServerUpdate)
	}
	window.onhashchange = socketFromHash
	socketFromHash()

	initMasterSocket(update => {
		update.sort((a, b) => b.numberOfClients - a.numberOfClients)
		serverlist.servers = update
		if (!sock && update.length) {
			window.location.hash = update[0].address
		}
	})
}

init()