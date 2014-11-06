var playerAttributes = [
	{
		"idString": "frags",
		"displayName": "Frags",
	},
	{
		"idString": "deaths",
		"displayName": "Deaths"
	}, {
		"idString": "accuracy",
		"displayName": "Accuracy",
		"valueDisplayModifierFunc": function(value) {return value + "%";}
	},{
		"idString": "name",
		"displayName": "Name"
	}, {
		"idString": "cn",
		"displayName": "CN"
	}
];

var teams = [];

// applys a modifier, if one is defined for the specified attribute, and makes the value HTML-safe
function finalizeValue(attribute, value) {
	var finalValue = value;

	if ("valueDisplayModifierFunc" in attribute) {
		finalValue = attribute.valueDisplayModifierFunc(value);
	}

	return escapeHtml(finalValue);
}

// frags (descending), then deaths (ascending), then accuracy (descending)
function scoreboardSortingFunction (a, b) {
	if (a.frags == b.frags) {
		if (a.deaths == b.deaths) {
			return b.accuracy - a.accuracy;
		} else {
			return a.deaths - b.deaths;
		}
	} else {
		return b.frags - a.frags;
	}
}

function makeTeamTable(name) {
	var table = document.createElement("table");
	table.id = "team-" + name;

	var header = document.createElement("thead");

	// info row: name & score of team

	var teamInfoRow = document.createElement("tr");
	teamInfoRow.id = "teaminfo-" + name;

	header.appendChild(teamInfoRow);

	// column header for player listing

	var columnHeadersRow = document.createElement("tr");

	playerAttributes.forEach(function (attribute) {
		var columnHeader = document.createElement("th");
		columnHeader.innerHTML = escapeHtml(attribute.displayName);
		columnHeadersRow.appendChild(columnHeader);
	})

	header.appendChild(columnHeadersRow);
	table.appendChild(header);

	var body = document.createElement("tbody");
	body.id = "players-" + name;
	table.appendChild(body);

	return table;
}

function addTeamTableToScoreboard(table) {
	var needsNewTableRow = teams.length % 2 == 0;

	var tr = null;
	if (needsNewTableRow) {
		tr = document.createElement("tr");
		$("teams").appendChild(tr);
	} else {
		tr = $("teams").lastChild;
	}

	var td = document.createElement("td");
	td.appendChild(table);

	tr.appendChild(td);
}

function updateTeamInfo(name, score) {
	var teamInfoRow = $("teaminfo-"+name);
	if (!teamInfoRow) {
		addTeamTableToScoreboard(makeTeamTable(name));
		teams.push(name);
		teamInfoRow = $("teaminfo-"+name);
	}

	teamInfoRow.innerHTML = "";

	var teamInfoName = document.createElement("td");
	teamInfoName.innerHTML = name;
	teamInfoName.colSpan = parseInt((playerAttributes.length / 2) + (playerAttributes.length % 2 == 0 ? 0 :1)).toString();
	teamInfoRow.appendChild(teamInfoName);

	var teamInfoScore = document.createElement("td");
	teamInfoScore.innerHTML = score;
	teamInfoScore.colSpan = parseInt((playerAttributes.length / 2)).toString();
	teamInfoRow.appendChild(teamInfoScore);
}

function updatePlayerListing(tbody, data) {
	tbody.innerHTML = "";

	// sort data
	data.sort(scoreboardSortingFunction)

	// fill table body with data
	data.forEach(function (entry) {
		var entryRow = document.createElement("tr");

		playerAttributes.forEach(function (attribute) {
			var entryCell = document.createElement("td");
			entryCell.innerHTML = finalizeValue(attribute, entry[attribute.idString]);
			entryRow.appendChild(entryCell);
		});

		tbody.appendChild(entryRow);
	});
}

function updateAllPlayerListings(players) {
	var organizedPlayers = {};

	teams.forEach(function(team) {
		organizedPlayers[team] = [];
	});

	players.forEach(function(player) {
		organizedPlayers[player.team].push(player);
	});

	for (team in organizedPlayers) {
		updatePlayerListing($("players-"+team), organizedPlayers[team]);
	}
}
