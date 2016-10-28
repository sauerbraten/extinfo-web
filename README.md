# extinfo-web

Go web app that lets you track the state of a Cube 2: Sauerbraten game server.

Try it [here](http://extinfo.sauerworld.org/)!

## How it works

Viewers subscribe to Sauerbraten server state information by opening a websocket connection. When that particular game server is not tracked yet, the extinfo server spawns a poller that periodically (every 5 seconds) requests state information from the Sauerbraten server using the [extinfo package](https://github.com/sauerbraten/extinfo). The current state is then sent down to every viewer subscribing to that server via a hub managing publishers (game server pollers) and subscribers (websocket viewers consuming the state information).

Vue.js is used to display state information in the browser.