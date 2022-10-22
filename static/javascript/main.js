console.log(`yahooo`);

document.getElementById("mainInputBox").addEventListener("keyup", function(event) {
    const shelfUrl = document.getElementById("mainInputBox").value
    if (event.key === "Enter") {
        initWebsocketConn(shelfUrl)
        // https://www.goodreads.com/review/list/1753152-sharon?shelf=fantasy

    }
});

function initWebsocketConn(shelfURL) {
    const ws = new WebSocket(`ws://localhost:2945/ws?shelfurl=${encodeURIComponent(shelfURL)}`);

    ws.onopen = function(e) {
        createRequestStatusBox() 
    };

    ws.onmessage = function(event) {
        const msg = JSON.parse(event.data)
        console.log(msg)

        if (isNewBookFromGoodReads(msg)) {
            writeBook(msg.bookinfo)
        }

        console.log(msg.crawlStats)
        updateStats(msg.crawlStats)
    }

    ws.onclose = function(event) {
        if (event.wasClean) {
            console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
        } else {
            console.log('[close] Connection died');
        }
    };
}

function isNewBookFromGoodReads(msg) {
    return msg.bookinfo != undefined
}

function writeBook(book) {
    document.getElementById("goodReadsBooksCol").innerHTML += `
    <div class="row goodReadsBookBox mt-2">
        <div class="col-1 text-center">
                <img
                    src="${book.cover}"
                    style="width: 3rem"
                >
        </div>
        <div class="col-4">
            <div class="row bookTitle" >
                ${book.title}
            </div>
            <div class="row bookSeriesText">
                ${book.seriesText}
            </div>
            <div class="row bookAuthor">
                ${book.author}
            </div>
        </div>
        <div class="col-4">
        
        </div>
        <div class="col-1">
            <img 
                src="static/images/icons8-tick-box.svg"
                style="width: 4.3rem; filter: sepia(60%)"
                class="tickIconDefault"
            >
        </div>
    </div>
    `
}

function updateStats(crawlStats) {
    document.getElementById("crawlInfoBox").textContent = `${crawlStats.booksCrawled}/${crawlStats.totalBooks}`
}