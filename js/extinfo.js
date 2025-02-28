import { scoreboard, serverlist, resetScoreboard } from './model.js'
import { initServerSocket, initMasterSocket, free } from './sockets.js'
import { renderUI } from './ui.js'
import { names } from './names.js'
import { initServerSocket } from './web_sockets.js'

let sock = null

const init = () => {
	if (!('WebSocket' in window)) {
		alert('sorry, but your browser does not support websockets', 'try updating your browser')
		return
	}

	renderUI(scoreboard, serverlist)

	const socketFromHash = () => {
		if (!window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
			return
		}
		free(sock)
		resetScoreboard()
		document.title = 'loading … – extinfo'
		window.scrollTo(window.scrollLeft, 0)
		sock = initServerSocket(window.location.hash.substring(1), processServerUpdate)
	}
	window.onhashchange = socketFromHash
	socketFromHash()

	initMasterSocket(processMasterServerUpdate)
}

const processMasterServerUpdate = (update) => {
	update.sort((a, b) => b.num_clients - a.num_clients)
	serverlist.servers = update
	renderUI(scoreboard, serverlist)
	if (!sock && update.length) {
		window.location.hash = update[0].address
	}
}

const processServerUpdate = (update) => {
	resetScoreboard()

	scoreboard.info = update.serverinfo
	document.title = `${update.serverinfo.description} – extinfo`

	for (const name in update.teams) {
		scoreboard.teams[name] = {...update.teams[name], players: []}
	}

	for (const player of Object.values(update.players)) {
		if (names.state(player.state) == 'spectator') {
			scoreboard.spectators.push(player)
		} else {
			scoreboard.teams[player.team]?.players.push(player) || scoreboard.teamless.push(player)
		}
	}

	renderUI(scoreboard, serverlist)
}

init()
