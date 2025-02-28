import {html, nothing, render} from 'https://cdn.jsdelivr.net/npm/lit-html@3/+esm'
import {map} from 'https://cdn.jsdelivr.net/npm/lit-html@3/directives/map/+esm'
import {styleMap} from 'https://cdn.jsdelivr.net/npm/lit-html@3/directives/style-map/+esm'
import {ifDefined} from 'https://cdn.jsdelivr.net/npm/lit-html@3/directives/if-defined/+esm'
import {names} from './names.js'

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
    const byStats = (a, b) => {
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
    }

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
                    ${map(players.sort(byStats), p => html`
                    <tr>
                        <td class='count'>${p.frags}</td>
                        <td class='count'>${p.deaths}</td>
                        <td class='count'>${p.accuracy}%</td>
                        <td>${playerName(p.name, p.cn, names.priv(p.privilege))}</td>
                    </tr>`)}
                </tbody>
            </table>
        </div>
    </div>`
}

const scoreboardComponent = ({info, teams, teamless, spectators}) => {
    const bgImgOverlayCSS = 'linear-gradient(rgba(0, 0, 0, .3), rgba(0, 0, 0, .3))'
    const bgImgMapshotCSS = m => `url('//sauertracker.net/images/mapshots/${m}.jpg') no-repeat center center / cover`
    const bgImgFallbackCSS = bgImgMapshotCSS('firstevermap')
    const backgroundImageCSS = (m) => `${bgImgOverlayCSS}, ${bgImgMapshotCSS(m)}, ${bgImgFallbackCSS}`

    return html`
    <main id='scoreboard' style='${styleMap({ background: backgroundImageCSS(info.map) })}'>
		<header>
			<h1 class='scrollable-x'>${info.description}</h1>
			<h3 class='scrollable-x'>
				<strong>${names.gm(info.game_mode)}</strong> &nbsp; on &nbsp; <strong>${info.map}</strong>
				<br>
				${timeRemaining(info.secs_left)}${info.paused ? ' &nbsp; | &nbsp; paused' : nothing } &nbsp; | &nbsp; ${names.mm(info.master_mode)} &nbsp; | &nbsp; ${info.num_clients}/${info.num_slots}
			</h3>
		</header>

		<section class='flex flex-row'>
            ${map(Object.values(teams), t => playerList(`${t.name}: ${t.score}`, t.players))}
			${teamless.length ? playerList('players', teamless) : nothing}
		</section>

        ${spectators.length==0 ? nothing : html`
        <section>
			<h2>spectators</h2>
			<div class='flex flex-row centered'>
				${map(spectators, s => playerName(s.name, s.cn, names.priv(s.privilege)))}
			</div>
		</section>`}
	</main>`
}

const serverlistComponent = ({servers}) => {
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
                        <td class='count'>${s.num_clients}</td>
                        <td>
                            <a href='${'#'+s.address}' class='subtle' title='${ifDefined(s.mod)}'>${s.description}</a>
                        </td>
                        <td>${names.gm(s.game_mode)}</td>
                        <td>${s.map}</td>
                        <td class='centered'>${timeRemaining(s.secs_left)}</td>
                        <td>${names.mm(s.master_mode)}</td>
                    </tr>`)}
                </tbody>
            </table>
        </div>`
      : html`<div><p class='centered'>loading...</p></div>`
}

const body = (scoreboard, serverlist) => html`
    ${scoreboardComponent(scoreboard)}

    <aside id='serverlist' class='flex flex-col'>
        <h2>other servers</h2>
        ${serverlistComponent(serverlist)}
    </aside>`

export const renderUI = (scoreboard, serverlist) => {
    render(body(scoreboard, serverlist), document.body, {renderBefore: document.getElementById('footer')})
}
