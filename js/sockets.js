function init(path, onUpdate) {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const url = `${protocol}//${window.location.host}${path}`
    const sock = new WebSocket(url)
    window.addEventListener('beforeunload', () => sock.close())
    sock.onerror = () => alert(`could not connect to a websocket server at ${url}!`)
    sock.onmessage = (msgEvent) => { onUpdate(JSON.parse(msgEvent.data)) }
    return sock
}

export const initSocket = (addr, onUpdate) => init(`/server/${addr}`, onUpdate)

export const initMasterSocket = (onUpdate) => init('/master', onUpdate)

export function free(sock) {
    if (!sock) {
        return
    }
    sock.close()
    sock.onerror = null
    sock.onmessage = null
    sock = null
}