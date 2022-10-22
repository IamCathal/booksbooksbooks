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
    };

    ws.onmessage = function(event) {
        const msg = JSON.parse(event.data)
        console.log(msg)

        if (isNewBookFromGoodReads(msg)) {
            writeBook(msg.bookinfo)
        }
        if (isSearchResult(msg)) {
            fillInSearchResult(msg.searchResult)
        }

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

function isSearchResult(msg) {
    return msg.searchResult != undefined
}

function writeBook(book) {
    document.getElementById("goodReadsBooksCol").innerHTML += `
    <div class="row goodReadsBookBox mt-2" id="${book.id}-goodreadsInfo">
        <div class="col-1 text-center">
                <a href="${book.link}">
                    <img
                        src="${book.cover}"
                        style="width: 3rem"
                    >
                </a>
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
        <div class="col" id="${book.id}-theBookshopResults">

        </div>
    </div>
    `
}

function fillInSearchResult(msg) {
    console.log(msg)
    if (msg.titleMatches.length > 1) {
        console.debug(msg)
    }
    if (msg.titleMatches.length == 0) {
        msg.titleMatches[0] = {
            "title":"",
            "author":"",
            "price":"",
            "link":"",
            "cover":""
        }
    }

    document.getElementById(`${msg.searchBook.id}-theBookshopResults`).innerHTML = `
                    <div class="row">
                        <div class="col"style="border: 2px solid red" >
                            <div class="row justify-content-md-center titleMatch" style="border: 1px dashed black">
                                Possible Match
                            </div>
                            <div class="row">
                                <div class="col searchResultBook" style="border: 1px dotted black">
                                    <div class="row">
                                        <div class="col-3 pl-2" style="border: 1px solid blue">
                                            <a href="${msg.titleMatches[0].link}">
                                                <img
                                                    src="${msg.titleMatches[0].cover}"
                                                    style="width: 3rem"
                                                >
                                            </a>
                                        </div>
                                        <div class="col" style="border: 1px solid green">
                                            <div class="row" style="font-weight: bold; font-size: 0.8rem">
                                            ${msg.titleMatches[0].title}
                                            </div>
                                            <div class="row" style="font-size: 0.6rem">
                                            
                                            </div>
                                            <div class="row" style="font-size: 0.6rem">
                                                ${msg.titleMatches[0].author}
                                            </div>
                                            <div class="row" style="font-weight: bold; font-size: 0.7rem">
                                                ${msg.titleMatches[0].price}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div class="col"style="border: 2px solid purple" >
                            <div class="row justify-content-md-center authorMatch" style="border: 1px dashed black">
                                Other books from author
                            </div>
                            <div class="row authorMatches">
                                ${generateHTMLForAuthorMatches(msg.authorMatches)}
                            </div>
                        </div>
                    </div>
    `
}

function generateHTMLForAuthorMatches(authorMatches) {
    let resultHTML = "";
    const numMatches = authorMatches.length

    for (let i = 0; i < authorMatches.length; i++) {
        const leftPosition = (100 / numMatches) * i

        resultHTML += `
        <a href="${authorMatches[i].link}">
            <img
                src="${authorMatches[i].cover}"
                style="position: absolute; width: 3rem"
            >
        </a>
        `
    }
    return resultHTML
}

function updateStats(crawlStats) {

    document.getElementById("statsBookFound").textContent = crawlStats.totalBooks
    document.getElementById("statsBooksCrawled").textContent = crawlStats.booksCrawled
    document.getElementById("statsBooksSearched").textContent = crawlStats.booksSearched
    document.getElementById("statsBookMatchesFound").textContent = crawlStats.bookMatchFound

    document.getElementById("crawlProgressBarSpanID").style.width = `${Math.floor((crawlStats.booksSearched/crawlStats.totalBooks)*100)}%`
    console.log(document.getElementById("crawlProgressBarSpanID").style.width)
}