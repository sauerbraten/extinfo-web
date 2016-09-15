var sock
var port = '28785'
var host = '164.132.110.240'

var serverinfo = {}
var teams = {}
var players = {}

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
		console.log('received:', m.data)
		var parts = m.data.split('\t')

		var field = parts[0]

		switch (field) {
			case 'serverinfo':
				serverinfo = JSON.parse(parts[1])
				return

			case 'team':
				var team = JSON.parse(parts[1])

				if (team == 'delete') {
					delete teams[parts[2]]
				} else {
					teams[team.name] = team
				}
				return

			case 'player':
				var player = JSON.parse(parts[1])
				if (player == 'delete') {
					delete players[parts[2]]
				} else {
					players[players.cn] = player
				}
				return

			case 'error':
				var err = JSON.parse(parts[1])
				error(err.message, error.hint)
				sock.close()
				return
		}
	}
}

function error(err) {
	console.log(err);
}