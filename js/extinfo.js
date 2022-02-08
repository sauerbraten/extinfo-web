import { scoreboard, serverlist, resetScoreboard } from './model.js'
import { initSocket, initMasterSocket, free } from './sockets.js'
import { render } from 'https://unpkg.com/lit-html?module'
import { ui } from './ui.js'

let sock = null

function init() {
	if (!('WebSocket' in window)) {
		alert('sorry, but your browser does not support websockets', 'try updating your browser')
		return
	}

	updateUI()

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

	initMasterSocket(processMasterServerUpdate)
}

function processMasterServerUpdate(update) {
	update.sort((a, b) => b.numberOfClients - a.numberOfClients)
	serverlist.servers = update
	updateUI()
	if (!sock && update.length) {
		window.location.hash = update[0].address
	}
}

function processServerUpdate(update) {
	console.log('received current server update')

	scoreboard.info = update.serverinfo
	document.title = update.serverinfo.description + ' – extinfo'

	let teams = new Map(), teamless = [], spectators = []

	if ('teams' in update) {
		for (const teamName in update.teams) {
			let team = update.teams[teamName]
			team.players = []
			teams.set(teamName, team)
		}
	}

	for (const cn in update.players) {
		let player = update.players[cn]
		if (player.state == 'spectator') {
			spectators.push(player)
		} else if (teams.has(player.team)) {
			teams.get(player.team).players.push(player)
		} else {
			teamless.push(player)
		}
	}

	if (teams.size) {
		scoreboard.teamless = []
		scoreboard.teams = teams
	} else {
		scoreboard.teams.clear()
		scoreboard.teamless = teamless
	}
	scoreboard.spectators = spectators

	updateUI()
}

function updateUI() {
	render(ui(scoreboard, serverlist), document.body, {renderBefore: document.getElementById('footer')})
}

init()