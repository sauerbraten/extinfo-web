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
			console.log(" - socket closed - ");
		};

		sock.onmessage = function(m) {
			console.log("received:", m.data);
			var parts = m.data.split("\t");

			var field = parts[0];
			var value = parts[1].replace(/</g, "&lt;").replace(/>/g, "&gt;");

			if (field === "timeleft") {
				var minutes = Math.floor(value/60);
				var seconds = value % 60;
				value = (minutes < 10 ? '0' : '') + minutes + ":" + (seconds < 10 ? '0' : '') + seconds;
			}

			$("extinfo-{{.Id}}-" + field).innerHTML = value.replace(/</g, "&lt;").replace(/>/g, "&gt;");
		};

		sock.onopen = function(e) {
			console.log(" - socket opened - ");
			sock.send(addr + "\t" + port);
			console.log("sent:    ", addr + "\t" + port);
		};
	}

	$("extinfo-{{.Id}}").innerHTML = "<table><tbody><tr><td id='extinfo-{{.Id}}-description' colspan='2'>&nbsp;</td></tr><tr><td>Game Mode</td><td id='extinfo-{{.Id}}-gamemode'></td></tr><tr><td>Map</td><td id='extinfo-{{.Id}}-map'></td></tr><tr><td>Clients</td><td><span id='extinfo-{{.Id}}-numberofclients'></span>/<span id='extinfo-{{.Id}}-maxnumberofclients'></span></td></tr><tr><td>Master Mode</td><td id='extinfo-{{.Id}}-mastermode'></td></tr><tr><td>Time Left</td><td><span id='extinfo-{{.Id}}-timeleft'></span> minutes</td></tr></tbody></table>";
	updatesocket();
})()