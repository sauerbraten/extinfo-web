(function() {
	var $ = function (id) {
		return document.getElementById(id);
	};

	var sock;
	var addr = "{{.Addr}}";
	var port = "{{.Port}}";

	function updatesocket() {
		sock = new WebSocket("ws://{{.Host}}/ws");

		sock.onclose = function(e) {
			//console.log(" - socket closed - ");
		};

		sock.onmessage = function(m) {
			console.log("received:", m.data);
			var parts = m.data.split("\t");

			var field = parts[0];
			var value = parts[1].replace(/</g, "&lt;").replace(/>/g, "&gt;");

			switch (field) {
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

			$("extinfo-{{.Id}}-" + field).innerHTML = value.replace(/</g, "&lt;").replace(/>/g, "&gt;");
		};

		sock.onopen = function(e) {
			//console.log(" - socket opened - ");
			sock.send(addr + ":" + port);
			//console.log("sent:    ", addr + ":" + port);
		};
	}

	function error(errorMessage, fixTip) {
		var error = document.createElement("div");
		error.className = "extinfo-error";

		var message = document.createElement("p");
		message.innerHTML = errorMessage;
		error.appendChild(message);

		var tip = document.createElement("small");
		tip.innerHTML = fixTip;
		error.appendChild(tip);

		$("extinfo-{{.Id}}").innerHTML = "";
		$("extinfo-{{.Id}}").appendChild(error);
	}

	if (!('WebSocket' in window)) {
		error("sorry, but your browser does not support websockets", "try updating your browser");
		return;
	}

	$("extinfo-{{.Id}}").innerHTML = "<table><tbody><tr><td id='extinfo-{{.Id}}-description' colspan='2'>&nbsp;</td></tr><tr><td>Game Mode</td><td id='extinfo-{{.Id}}-gamemode'></td></tr><tr><td>Map</td><td id='extinfo-{{.Id}}-map'></td></tr><tr><td>Clients</td><td><span id='extinfo-{{.Id}}-numberofclients'></span>/<span id='extinfo-{{.Id}}-maxnumberofclients'></span></td></tr><tr><td>Master Mode</td><td id='extinfo-{{.Id}}-mastermode'></td></tr><tr><td>Time Left</td><td id='extinfo-{{.Id}}-timeleft'></td></tr></tbody></table>";

	updatesocket();
})()