console.log(`yahooo`);

const ws = new WebSocket(`ws://localhost:2945/ws`);

ws.onopen = function(e) {
    createRequestStatusBox() 
};

ws.onmessage = function(event) {
    console.log(event)
}

ws.onclose = function(event) {
    if (event.wasClean) {
        console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        console.log('[close] Connection died');
    }
    hideCreateStatusRequestBox() 
};