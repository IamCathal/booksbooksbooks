giveSwayaaangBordersToItems()

getAndRenderSeriesCrawlIfPossible()

function initWsConnection() {
    let booksFound = 0
    const ws = new WebSocket(`ws://${getCurrentHostname()}/ws/seriescrawl`);

    ws.onopen = function(e) {
    };

    ws.onmessage = function(event) {
        const msg = JSON.parse(event.data)
        console.log(msg)

        if (isErrorMsg(msg)) {
            console.error(`Error returned from backend: ${msg.error}`)
        }

        if (isNewSeriesFoundMessage(msg)) {
            renderSeriesForCurrentCrawl(msg.series)
        }

        if (isSearchResultMesage(msg)) {
            renderSearchResult(msg.searchBook, msg.match)
        }

        // if (isNewBookFromGoodReads(msg)) {
        //     writeBook(msg.bookinfo)
        //     allBooks.push({
        //         "sequentialID": booksFound,
        //         "bookInfo": msg.bookinfo,
        //         "titleMatches": {},
        //         "authorMatches": {}
        //     })
        //     booksFound++
        // }

        // if (isSearchResult(msg)) {
        //     fillInSearchResult(msg.searchResult)
        //     addSearchResultsToBookArr(msg.searchResult, allBooks)
        // }

        // if (isNewBookAvailable(msg)) {
        //     renderAndAddBookToNewAvailableBookList(msg.newAvailableBook)
        // }

        updateStats(msg.seriesCrawlStats)
    }

    ws.onclose = function(event) {
        if (event.wasClean) {
            console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
        } else {
            console.log('Connection was closed by backend');
            document.getElementById("crawlProgressDiv").style.display = 'none';
        }
    };
}

document.getElementById("seriesCrawlButtonElem").addEventListener("click", (ev) => {
    document.getElementById("seriesRowsOutputDiv").innerHTML = ""
    showCrawlInfoElements()
    initWsConnection()
})

function showCrawlInfoElements() {
    document.getElementById("crawlInfoBox").style.display = 'inline'
    document.getElementById("crawlProgressDiv").style.backgroundColor = '#e1e4e8';
}


function isNewSeriesFoundMessage(msg) {
    return msg.series != undefined
}


function isSearchResultMesage(msg) {
    return msg.match != undefined
}

function hasCrawlStatsInMessage(msg) {
    return msg.seriesCrawlStats != undefined
}

function isErrorMsg(msg) {
    return msg.error != undefined
}

function getCurrentHostname() {
    return new URL(window.location.href).host
}

function getAndRenderSeriesCrawlIfPossible() {
    getSeriesCrawl().then((seriesCrawl) => {
        seriesCrawl.forEach(series => {
            renderSeriesForExistingCrawl(series)
        })
    }, (err) => {
        console.error(seriesCrawl)
    })
}

function renderSeriesForCurrentCrawl(seriesInfo) {
    document.getElementById("seriesRowsOutputDiv").innerHTML += 
    `
    <div class="row thinBorderBox mb-2 pl-2 pr-2 pt-1 pb-1">
        <div class="col">
            <div class="row">
                <div class="col pl-0" style="font-weight: bold; font-size: 1.2rem">
                    <a href="${seriesInfo.link}">${seriesInfo.title}</a>
                </div>
            </div>
            <div class="row">
                <div class="col pl-0">
                    ${seriesInfo.author} - ${seriesInfo.primaryWorks} primary and ${seriesInfo.totalWorks} total works
                </div>
            </div>
            <div class="row" style="width: 100%;">
                ${renderBooksInSeriesForCurrentCrawl(seriesInfo.books)}
            </div>
        </div>
    </div>
    `

}
function renderSeriesForExistingCrawl(seriesInfo) {
    document.getElementById("seriesRowsOutputDiv").innerHTML += 
    `
    <div class="row thinBorderBox mb-2 pl-2 pr-2 pt-1 pb-1">
        <div class="col">
            <div class="row">
                <div class="col pl-0" style="font-weight: bold; font-size: 1.2rem">
                    <a href="${seriesInfo.link}">${seriesInfo.title}</a>
                </div>
            </div>
            <div class="row">
                <div class="col pl-0">
                    ${seriesInfo.author} - ${seriesInfo.primaryWorks} primary and ${seriesInfo.totalWorks} total works
                </div>
            </div>
            <div class="row" style="width: 100%;">
                ${renderSeriesBooksForExistingCrawl(seriesInfo.books)}
            </div>
        </div>
    </div>
    `
}

function renderSeriesBooksForExistingCrawl(books) {
    let output = ""

    books.forEach(book => {
        output += 
        `
        <div class="col-3 pt-2 pb-2" style="height: 6.8rem">
            <div class="row">
                <div class="col-3 pl-2">
                    <img src="${book.bookInfo.cover}" style="width: 3.75rem; height: 5.7rem;">
                </div>
                <div class="col-9" style="height: 6rem">
                    <div class="row">
                        <div class="col">
                            <p class="mb-0 seriesBookTitleOverflow" style="font-weight: bold" title="${book.bookInfo.title}"> ${book.bookInfo.title} </p>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col" style="font-size: 0.7rem">
                            ${book.bookSeriesText}
                        </div>
                    </div>
                    <div class="row">
                        <div class="col mt-0" style="font-size: 0.7rem">
                            ${book.bookInfo.rating} stars. Published ${book.bookInfo.published}
                        </div>
                    </div>
                    <div class="row pl-4" style="position: absolute; bottom: 0.5rem; left: 0;  width: 100%">
                        <div class="col-11 mt-0 thinBorderBox text-center" style="font-size: 0.7rem; ${hasSearchResult(book.theBookshopMatch) ? "" : "border-color: #74797b; color: #74797b"}" id="${book.bookInfo.id}-theBookshopResults">
                            ${getSearchResultText(book.theBookshopMatch)}
                        </div>
                    </div>
                </div>
            </div>
        </div>
        `
    })
    return output
}

function renderBooksInSeriesForCurrentCrawl(books) {
    let output = ""

    books.forEach(book => {
        output += 
        `
        <div class="col-3 pt-2 pb-2" style="height: 6.8rem">
            <div class="row">
                <div class="col-3 pl-2">
                    <img src="${book.bookInfo.cover}" style="width: 3.75rem; height: 5.7rem;">
                </div>
                <div class="col-9" style="height: 6rem">
                    <div class="row">
                        <div class="col">
                            <p class="mb-0 seriesBookTitleOverflow" style="font-weight: bold" title="${book.bookInfo.title}"> ${book.bookInfo.title} </p>
                        </div>
                    </div>
                    <div class="row">
                        <div class="col" style="font-size: 0.7rem">
                            ${book.bookSeriesText}
                        </div>
                    </div>
                    <div class="row">
                        <div class="col mt-0" style="font-size: 0.7rem">
                            ${book.bookInfo.rating} stars. Published in ${book.bookInfo.published}
                        </div>
                    </div>
                    <div class="row pl-4" style="position: absolute; bottom: 0.5rem; left: 0;  width: 100%">
                        <div class="col-11 mt-0 thinBorderBox text-center ${hasSearchResult(book.theBookshopMatch) ? "" : "skeleton"}" style="font-size: 0.7rem; ${hasSearchResult(book.theBookshopMatch) ? "" : "border-color: #74797b; color: #74797b"}" id="${book.bookInfo.id}-theBookshopResults">
                            ${getSearchResultText(book.theBookshopMatch)}
                        </div>
                    </div>
                </div>
            </div>
        </div>
        `
    })
    return output
}

function renderSearchResult(searchBook, searchResult) {
    document.getElementById(`${searchBook.id}-theBookshopResults`).innerHTML = getSearchResultText(searchResult)
    document.getElementById(`${searchBook.id}-theBookshopResults`).classList.remove("skeleton")
}

function getSearchResultText(searchResult) {
    if (hasSearchResult(searchResult)) {
        return `<a style="font-weight: bold" href="${searchResult.link}">${searchResult.price} match found</a>`
    } else {
        return `<p class="mb-0 mt-0" style="color: #22242f">.</p>`
    }
}

function hasSearchResult(searchResult) {
    return searchResult.title != ""
}

function updateStats(crawlStats) {
    document.getElementById("statsSeriesFound").textContent = crawlStats.seriesCount
    if (crawlStats.totalBooksInSeries == -1) {
        crawlStats.totalBooksInSeries = 0
    }
    document.getElementById("statsBooksInAllSeries").textContent = crawlStats.totalBooksInSeries
    document.getElementById("statsNotAvailable").textContent = crawlStats.booksSearchedOnTheBookshop - crawlStats.bookMatchesFound
    document.getElementById("statsAvailable").textContent = crawlStats.bookMatchesFound

    document.getElementById("crawlProgressBarSpanID").style.width = `${Math.floor((crawlStats.booksSearchedOnTheBookshop/crawlStats.totalBooksInSeries)*100)}%`
}

function giveSwayaaangBordersToItems() {
    document.getElementById("availableLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("seriesLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("shelfLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("settingsLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("seriesCrawlButtonElem").style = swayaaangBorders(0.8)
}

function swayaaangBorders(borderRadius) {
    const borderArr = [
        `border-top-right-radius: ${borderRadius}rem;`, 
        `border-bottom-right-radius: ${borderRadius}rem;`,
        `border-top-left-radius: ${borderRadius}rem;`,
        `border-bottom-left-radius: ${borderRadius}rem;`,
    ]

    let borderRadiuses = "";
    for (let k = 0; k < 4; k++) {
        const randNum = Math.floor(Math.random() * 2)
        if (randNum % 2 == 0) {
            borderRadiuses += borderArr[k]
        }
    } 
    return borderRadiuses
}

function getSeriesCrawl() {
    return new Promise((resolve, reject) => {
        fetch(`/getseriescrawl`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res)
        }, (err) => {
            reject(err)
        });
    })
}