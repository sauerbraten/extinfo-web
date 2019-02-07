import { scoreboard, serverlist } from './model.js'

Vue.component('time-remaining', {
    props: {
        secsLeft: {
            type: Number,
            default: 0
        }
    },
    data() {
        return {
            liveSecsLeft: 0,
            secsLeftUpdaters: []
        }
    },
    watch: {
        secsLeft: {
            handler: function (newSecsLeft) {
                for (const updater of this.secsLeftUpdaters) {
                    window.clearTimeout(updater)
                }
                this.secsLeftUpdaters = []
                this.liveSecsLeft = newSecsLeft
                for (let i = 1; i < 5 && i <= newSecsLeft; i++) {
                    this.secsLeftUpdaters.push(window.setTimeout(() => this.liveSecsLeft--, i * 1000))
                }
            },
            immediate: true
        }
    },
    computed: {
        timeLeft() {
            const pad = i => (i < 10 ? '0' : '') + i
            const formatTimeLeft = s => `${pad(Math.floor(s / 60))}:${pad(s % 60)}`
            return formatTimeLeft(this.liveSecsLeft)
        }
    },
    template: `<span>{{timeLeft}}</span>`
})

Vue.component('player-name', {
    props: {
        player: {
            type: Object,
            default: {
                name: '',
                clientNum: -1,
                privilege: 'none'
            }
        }
    },
    template: `
            <span>
                <span v-if='player.privilege != "none"' :class='"priv-"+player.privilege' :title='player.privilege'>{{player.name}}</span>
                <span v-else>{{player.name}}</span>
                <span class='cn'>({{player.clientNum}})</span>
            </span>`,
})

Vue.component('player-list', {
    props: {
        title: {
            type: String,
            default: "players"
        },
        players: {
            type: Array,
            default: []
        }
    },
    computed: {
        sortedPlayers() {
            return this.players.sort((a, b) => {
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
        }
    },
    template: `
            <div class='team flex flex-col'>
                <h2>{{title}}</h2>
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
                            <tr v-for='player in sortedPlayers'>
                                <td class='count'>{{player.frags}}</td>
                                <td class='count'>{{player.deaths}}</td>
                                <td class='count'>{{player.accuracy}}%</td>
                                <td><player-name :player='player'></player-name></td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>`,
})

Vue.component('server-list', {
    props: {
        servers: {
            type: Array,
            default: []
        }
    },
    template: `
            <div class='scrollable-x'>
                <table v-if='servers.length'>
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
                        <tr v-for='server in servers'>
                            <td class='count'>{{server.numberOfClients}}</td>
                            <td>
                                <a :href='"#"+server.address' class='subtle'>{{server.description}}</a>
                            </td>
                            <td>{{server.gameMode}}</td>
                            <td>{{server.map}}</td>
                            <td class='centered'><time-remaining :secs-left='server.secsLeft'></time-remaining></td>
                            <td>{{server.masterMode}}</td>
                        </tr>
                    </tbody>
                </table>
                <p v-else class='centered'>
                    loading...
                </p>
            </div>`
})

function ui() {
    const bgImgOverlayCSS = 'linear-gradient(rgba(0, 0, 0, .3), rgba(0, 0, 0, .3))'
    const bgImgMapshotCSS = map => `url('//sauertracker.net/images/mapshots/${map}.jpg') no-repeat center center / cover`
    const bgImgFallbackCSS = bgImgMapshotCSS('firstevermap')

    new Vue({
        el: '#scoreboard',
        data: scoreboard,
        computed: {
            backgroundImageCSS: () => `${bgImgOverlayCSS}, ${bgImgMapshotCSS(scoreboard.info.map)}, ${bgImgFallbackCSS}`
        }
    })

    new Vue({
        el: '#serverlist',
        data: serverlist
    })
}

ui()