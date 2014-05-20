var $ = function (id) {
	return document.getElementById(id);
};

var sock;
var port = "";
var host = "";

function updatesocket() {
	var newhost = $("host").innerHTML.trim().replace(/<br ?\/?>/, "");
	var newport = $("port").innerHTML.trim().replace(/<br ?\/?>/, "");

	if (newhost == host && newport == port) {
		return;
	}

	if (typeof(sock) != "undefined") {
		sock.close();
		sock = null;
	}

	sock = new WebSocket("ws://"+window.location.host+"/ws");

	sock.onopen = function(e) {
		clearTable();
		$("description").innerHTML = "connecting...";
		console.log(" - socket opened - ");
		sock.send(newhost + ":" + newport);
		console.log("    sent:", newhost + ":" + newport);
	};

	sock.onclose = function(e) {
		console.log(" - socket closed - ");
		document.title = "extinfo-web";
		$("heading").innerHTML = "not connected // extinfo-web";
	};

	sock.onmessage = function(m) {
		console.log("received:", m.data);
		var parts = m.data.split("\t");

		var field = parts[0];
		var value = parts[1].trim().replace(/</g, "&lt;").replace(/>/g, "&gt;");

		switch (field) {
		case "description":
			document.title = value + " // extinfo-web";
			$("heading").innerHTML = value + " // extinfo-web";
			break;

		case "maxnumberofclients":
			value = "/" + value;
			break;

		case "timeleft":
			if (value == 0) {
				value = "intermission";
				break;
			}

			var minutes = Math.floor(value/60);
			var seconds = value % 60;
			value = (minutes < 10 ? '0' : '') + minutes + ":" + (seconds < 10 ? '0' : '') + seconds + " minutes";
			break;

		case "error":
			var errorMessage = value;
			var fixTip = parts[2].trim().replace(/</g, "&lt;").replace(/>/g, "&gt;");
			error(errorMessage, fixTip);
			sock.close();
			return;
		}

		$(field).innerHTML = value;
	};

	host = newhost;
	port = newport;
	window.location.hash = "#" + host + ":" + port;
}

function init() {
	if (!('WebSocket' in window)) {
		error("sorry, but your browser does not support websockets", "try updating your browser");
		return;
	}

	if (window.location.hash.substring(1).match(/[a-zA-Z0-9\\.]+:[0-9]+/)) {
		var parts = window.location.hash.substring(1).split(':');
		$("host").innerHTML = parts[0];
		$("port").innerHTML = parts[1];
	}

	updatesocket();
}

function error(errorMessage, fixTip) {
	var error = document.createElement("div");
	error.className = "error";

	var message = document.createElement("p");
	message.innerHTML = errorMessage;
	error.appendChild(message);

	var tip = document.createElement("small");
	tip.innerHTML = fixTip;
	error.appendChild(tip);

	$("description").innerHTML = "";
	$("description").appendChild(error);
}

function clearTable() {
	$("description").innerHTML = "";
	$("gamemode").innerHTML = "";
	$("map").innerHTML = "";
	$("numberofclients").innerHTML = "";
	$("maxnumberofclients").innerHTML = "";
	$("mastermode").innerHTML = "";
	$("timeleft").innerHTML = "";
}