import {html, nothing} from 'https://unpkg.com/lit-html?module'
import {map} from 'https://unpkg.com/lit-html/directives/map?module'
import {styleMap} from 'https://unpkg.com/lit-html/directives/style-map?module'
import {ifDefined} from 'https://unpkg.com/lit-html/directives/if-defined?module'

const timeRemaining = (secsLeft) => {
    const pad = i => (i < 10 ? '0' : '') + i
    const formatTimeLeft = s => `${pad(Math.floor(s / 60))}:${pad(s % 60)}`
    return html`<span>${formatTimeLeft(secsLeft)}</span>`
}

const playerName = (name, cn, priv) => html`
    <span>
        ${priv=='none'
          ? html`<span>${name}</span>`
          : html`<span class='${'priv-'+priv}' title='${priv == 'none' ? nothing : priv}'>${name}</span>`
        }
        <span class='cn'>(${cn})</span>
    </span>`

const playerList = (title, players) => {
    const sorted = players.sort((a, b) => {
            // sorts by frags (descending), then deaths (ascending), then accuracy (descending)
            if (a.frags !== b.frags) {
                return b.frags - a.frags
            } else {
                if (a.deaths !== b.deaths) {
                    return a.deaths - b.deaths
                } else {
                    return b.accuracy - a.accuracy
                }
            }
        })

    return html`
    <div class='team flex flex-col'>
        <h2>${title}</h2>
        <div class='team-table scrollable-x'>
            <table>
                <thead>
                    <tr>
                        <td>frags</td>
                        <td>deaths</td>
                        <td>accuracy</td>
                        <td>name</td>
                    </tr>
                </thead>
                <tbody>
                    ${map(sorted, p => html`
                    <tr>
                        <td class='count'>${p.frags}</td>
                        <td class='count'>${p.deaths}</td>
                        <td class='count'>${p.accuracy}%</td>
                        <td>${playerName(p.name, p.clientNum, p.privilege)}</td>
                    </tr>`)}
                </tbody>
            </table>
        </div>
    </div>`
}

const scoreBoard = (info, teams, teamless, spectators) => {
    const bgImgOverlayCSS = 'linear-gradient(rgba(0, 0, 0, .3), rgba(0, 0, 0, .3))'
    const bgImgMapshotCSS = m => `url('//sauertracker.net/images/mapshots/${m}.jpg') no-repeat center center / cover`
    const bgImgFallbackCSS = bgImgMapshotCSS('firstevermap')
    const backgroundImageCSS = (m) => `${bgImgOverlayCSS}, ${bgImgMapshotCSS(m)}, ${bgImgFallbackCSS}`
    
    return html`
    <main id='scoreboard' style='${styleMap({ background: backgroundImageCSS(info.map) })}'>
		<header>
			<h1 class='scrollable-x'>${info.description}</h1>
			<h3 class='scrollable-x'>
				<strong>${info.gameMode}</strong> &nbsp; on &nbsp; <strong>${info.map}</strong>
				<br>
				${timeRemaining(info.secsLeft)}${info.paused ? ' &nbsp; | &nbsp; paused' : nothing } &nbsp; | &nbsp; ${info.masterMode} &nbsp; | &nbsp; ${info.numberOfClients}/${info.maxNumberOfClients}
			</h3>
		</header>

		<section class='flex flex-row'>
            ${map(teams, ([_, t]) => playerList(`${t.name}: ${t.score}`, t.players))}
			${teamless.length ? playerList('players', teamless) : nothing}
		</section>

        ${spectators.length==0 ? nothing : html`
        <section>
			<h2>spectators</h2>
			<div class='flex flex-row centered'>
				${map(spectators, s => playerName(s.name, s.clientNum, s.privilege))}
			</div>
		</section>`}
	</main>`
}


const serverList = (servers) => {
    return servers.length
      ? html`<div class='scrollable-x'>
            <table>
                <thead>
                    <tr>
                        <td>players</td>
                        <td>description</td>
                        <td>mode</td>
                        <td>map</td>
                        <td>time left</td>
                        <td>master mode</td>
                    </tr>
                </thead>
                <tbody>
                    ${map(servers, s => html`
                    <tr>
                        <td class='count'>${s.numberOfClients}</td>
                        <td>
                            <a href='${'#'+s.address}' class='subtle' title='${ifDefined(s.mod)}'>${s.description}</a>
                        </td>
                        <td>${s.gameMode}</td>
                        <td>${s.map}</td>
                        <td class='centered'>${timeRemaining(s.secsLeft)}</td>
                        <td>${s.masterMode}</td>
                    </tr>`)}
                </tbody>
            </table>
        </div>`
      : html`<div><p class='centered'>loading...</p></div>`
}

export const ui = ({info, teams, teamless, spectators}, serverlist) => html`
    ${scoreBoard(info, teams, teamless, spectators)}
	<aside id='serverlist' class='flex flex-col'>
		<h2>other servers</h2>
		${serverList(serverlist.servers)}
	</aside>`
