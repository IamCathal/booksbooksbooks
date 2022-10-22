console.log(`yahooo`);

initWebsocketConn("https://www.goodreads.com/review/list/1753152-sharon?shelf=fantasy")

function initWebsocketConn(shelfURL) {
    const ws = new WebSocket(`ws://localhost:2945/ws?shelfurl=${encodeURIComponent(shelfURL)}`);

ws.onopen = function(e) {
    createRequestStatusBox() 
};

ws.onmessage = function(event) {
    console.log(event.data)
}

ws.onclose = function(event) {
    if (event.wasClean) {
        console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        console.log('[close] Connection died');
    }
};
}
