var $ = function (id) {
	return document.getElementById(id);
};

var sock;
var addr = "";
var port = "";

var teamOf = {};
var stateOf = {};

function updateSocket() {
	var newAddr = "";
	var newPort = "";

	if ($("addr") && $("port")) {
		newAddr = $("addr").innerHTML.replace(/<br ?\/?>/, "");
		newPort = $("port").innerHTML.replace(/<br ?\/?>/, "");

		if (newAddr == addr && newPort == port) {
			return;
		}
	} else {
		newAddr = addr;
		newPort = port;
	}

	if (typeof(sock) != "undefined") {
		sock.close();
		sock = null;
	}

	sock = new WebSocket("ws://" + host + "/ws/extended");

	sock.onclose = function(e) {
		console.log(" - socket closed - ");
	};

	sock.onmessage = function(m) {
		console.log("received:", m.data);
		var parts = m.data.split("\t");

		var field = parts[0];

		if (field == "playerstats") {
			processPlayerStats(parts[1], parts[2], parts[3]);
			return;
		}

		var value = parts[1].replace(/</g, "&lt;").replace(/>/g, "&gt;");

		switch (field) {
			case "description":
			document.title = value + " // extinfo-web";
			$("heading").innerHTML = "<a href='/'>extinfo-web</a> //" + value;
			return;

			case "timeleft":
			if (value == 0) {
				value = "intermission";
				break;
			}

			var minutes = Math.floor(value/60);
			var seconds = value % 60;
			value = (minutes < 10 ? '0' : '') + minutes + ":" + (seconds < 10 ? '0' : '') + seconds + " minutes left";
			break;
		}

		$(field).innerHTML = value;
	};

	sock.onopen = function(e) {
		console.log(" - socket opened - ");
		sock.send(newAddr + ":" + newPort);
		console.log("sent:    ", newAddr + ":" + newPort);
	};

	addr = newAddr;
	port = newPort;
	window.location.hash = "#" + addr + ":" + port;
}

function init(host) {
	window.host = host;

	if (!('WebSocket' in window)) {
		var error = document.createElement("div");
		error.className = "error";

		var message = document.createElement("p");
		message.innerHTML = "sorry, but your browser does not support websockets";
		error.appendChild(message);

		var tip = document.createElement("small");
		tip.innerHTML = "try updating your browser";
		error.appendChild(tip);

		$("table").parentNode.replaceChild(error, $("table"));
		return;
	}

	if (window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
		var parts = window.location.hash.substring(1).split(':');
		if ($("addr") && $("port")) {
			$("addr").innerHTML = parts[0];
			$("port").innerHTML = parts[1];
		} else {
			window.addr = parts[0];
			window.port = parts[1];
		}
	} else {
		if ($("addr") && $("port")) {
			$("addr").innerHTML = "sauerleague.org";
			$("port").innerHTML = "10000";
		} else {
			window.addr = "sauerleague.org";
			window.port = "10000";
		}
	}
	updateSocket();
}

function processPlayerStats(cn, field, value) {
	value = value.replace(/</g, "&lt;").replace(/>/g, "&gt;");

	switch (field) {
	case "disconnected":
		$("team-"+teamOf[cn]).removeChild($("player-"+cn));
		delete teamOf[cn];
		delete stateOf[cn];
		return;

	case "state":
		if (value === "spectator") {
			addSpectator(cn);
			stateOf[cn] = "spectator";
		} else {
			stateOf[cn] = "playing";
		}
		break;

	case "team":
		if (stateOf[cn] == "spectator") {
			break;
		}

		var newTeamTable = $("team-" + value);

		if (newTeamTable == null) {
			newTeamTable = initTeamTable(value);
		}

		if (teamOf.hasOwnProperty(cn)) {
			var oldTeamTable = $("team-"+teamOf[cn]);
			var playerRow = oldTeamTable.removeChild($("player-"+cn));
			newTeamTable.appendChild(playerRow);
		} else {
			initPlayer(cn, value);
		}
		break;

	case "frags":
	case "deaths":
		if (stateOf[cn] != "playing") {
			break;
		}

	default:
		$("player-"+cn+"-"+field).innerHTML = value;
	}
}

function initPlayer(cn, team) {
	teamOf[cn] = team;
	var teamTable = $("team-"+team);

	if (teamTable == null) {
		teamTable = initTeamTable(team);
	}

	var playerRow = document.createElement("div");
	playerRow.className = "table-row";
	playerRow.id = "player-"+cn;

	var playerCN = document.createElement("p");
	playerCN.className = "table-cell right";
	playerCN.id = "player-"+cn+"-cn";
	playerRow.appendChild(playerCN);

	var playerName = document.createElement("p");
	playerName.className = "table-cell";
	playerName.id = "player-"+cn+"-name";
	playerRow.appendChild(playerName);

	var playerFrags = document.createElement("p");
	playerFrags.className = "table-cell right";
	playerFrags.id = "player-"+cn+"-frags";
	playerRow.appendChild(playerFrags);

	var playerDeaths = document.createElement("p");
	playerDeaths.className = "table-cell right";
	playerDeaths.id = "player-"+cn+"-deaths";
	playerRow.appendChild(playerDeaths);

	teamTable.appendChild(playerRow);

	$("player-"+cn+"-cn").innerHTML = cn;

	return playerRow;
}

function initTeamTable(team) {
	var teamCell = document.createElement("div");
	teamCell.className = "table-cell thin-padding";

	var teamTable = document.createElement("div");
	teamTable.className = "table inner-table full-width";

	var tableHead = document.createElement("div");
	tableHead.className = "table-header";

	var tableHeadRow0 = document.createElement("div");
	tableHeadRow0.className = "table-row";

	var dummy = document.createElement("p");
	dummy.className = "table-cell";
	dummy.innerHTML = "&nbsp;";
	tableHeadRow0.appendChild(dummy);

	var teamName = document.createElement("p");
	teamName.className = "centered table-cell";
	teamName.innerHTML = team;
	tableHeadRow0.appendChild(teamName);

	var teamScore = document.createElement("p");
	teamScore.className = "centered table-cell";
	teamScore.innerHTML = "0";
	tableHeadRow0.appendChild(teamScore);

	dummy = document.createElement("p");
	dummy.className = "table-cell";
	dummy.innerHTML = "&nbsp;";
	tableHeadRow0.appendChild(dummy);

	tableHead.appendChild(tableHeadRow0);

	var tableHeadRow1 = document.createElement("div");
	tableHeadRow1.className = "table-row grey-text";

	var tableHeadCN = document.createElement("p");
	tableHeadCN.className = "table-cell right";
	tableHeadCN.innerHTML = "cn";
	tableHeadRow1.appendChild(tableHeadCN);

	var tableHeadName = document.createElement("p");
	tableHeadName.className = "table-cell";
	tableHeadName.innerHTML = "name";
	tableHeadRow1.appendChild(tableHeadName);

	var tableHeadFrags = document.createElement("p");
	tableHeadFrags.className = "table-cell right";
	tableHeadFrags.innerHTML = "frags";
	tableHeadRow1.appendChild(tableHeadFrags);

	var tableHeadDeaths = document.createElement("p");
	tableHeadDeaths.className = "table-cell right";
	tableHeadDeaths.innerHTML = "deaths";
	tableHeadRow1.appendChild(tableHeadDeaths);

	tableHead.appendChild(tableHeadRow1);

	teamTable.appendChild(tableHead);

	var tableBody = document.createElement("div");
	tableBody.className = "table-body";
	tableBody.id = "team-"+team;

	teamTable.appendChild(tableBody);

	teamCell.appendChild(teamTable);

	$("teams").appendChild(teamCell);

	return tableBody;
}

function addSpectator(cn) {
	var spectatorCell = document.createElement("div");
	spectatorCell.className = "table-cell";

	var spectatorCN = document.createElement("span");
	spectatorCN.className = "grey-text";
	spectatorCN.id = "player-"+cn+"-cn";
	spectatorCell.appendChild(spectatorCN);

	spectatorCell.innerHTML += " ";

	var spectatorName = document.createElement("span");
	spectatorName.id = "player-"+cn+"-name";
	spectatorCell.appendChild(spectatorName);

	$("spectators").appendChild(spectatorCell);

	$("player-"+cn+"-cn").innerHTML = cn;
}