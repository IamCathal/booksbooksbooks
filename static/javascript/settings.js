giveSwayaaangBordersToItems()

document.getElementById("shelfCheckURLInputBox").value = ""
getAndRenderSettings()

function getAndRenderSettings() {
    getAndRenderOwnedBooksShelfURL()
    getAndRenderDiscordWebhookURL()
    getAndRenderDiscordMessageFormatPreference()
    getAndRenderAutomatedCrawlTime()
    getAndRenderSendAlertsWhenBookNoLongerAvailable()
    getAndRenderSendAlertOnlyWhenFreeShippingKicksIn()
    getAndRenderOnlyConsiderEnglishBooks()
    getAndRenderAddMoreAuthorBooksToAvailableList()
    getAndRenderSeriesInAutomatedCrawl()
    getAndRenderKnownAuthorsList()

    getAndRenderShelvesToCrawl()
}

document.getElementById("settingsSetAutomatedCheckTimeButton").addEventListener("click", (ev) => {
    Timepicker.showPicker({
        onSubmit: (time) => {
            document.getElementById("automatedCheckTime").textContent = time.formatted()
            setAutomatedCrawlTime(time.formatted())
        }
    })
})

document.getElementById("addShelfToCrawl").addEventListener("click", (ev) => {

    document.getElementById("newShelfToCheckShelfPreviewRow").style.display = "none"
    document.getElementById("automatedCheckShelfURLCheckStatsBox").innerHTML = ""
    document.getElementById("shelfUrlAutomatedCheckStatsTextBox").textContent = ""

    document.getElementById("addShelfToCrawl").classList.add("skeleton")
    const shelfUrl = document.getElementById("shelfCheckURLInputBox").value

    addShelfToCrawl(shelfUrl).then(() => {
        document.getElementById("addShelfToCrawl").classList.remove("skeleton")
        getAndRenderShelvesToCrawl()
    }, (err) => {
        console.error(err)
        document.getElementById("addShelfToCrawl").classList.remove("skeleton")
        document.getElementById("newShelfToCheckShelfPreviewRow").style.display = "flex"
        document.getElementById("shelfUrlAutomatedCheckStatsTextBox").innerHTML +=
        `
            <div class="col text-center">
                <p class="pl-4 mb-0">That's not a valid Goodreads shelf URL. This is an example of a valid shelf URL:</p>
                <p class=" pl-4 pb-0"> <a href="https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read">https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read </a> </p>
            </div>
        `
    })

    // setAutomatedShelfCheckURL(shelfUrl).then((res) => {
    //     getPreviewForBookShelf(shelfUrl).then(bookPreview => {
    //         document.getElementById("newShelfToCheckShelfPreviewRow").style.display = "flex"
    
    //         document.getElementById("shelfUrlAutomatedCheckStatsTextBox").textContent = 
    //             `Found ${bookPreview.totalBooks} books. Should take roughly ${Math.floor(getSWAGEstimateForCrawlTime(bookPreview.totalBooks))}s to crawl`
    //         bookPreview.books.forEach(book => {
    //             document.getElementById("automatedCheckShelfURLCheckStatsBox").innerHTML += `
    //                             <div class="col-1 pr-1">
    //                                 <img 
    //                                     src="${book.cover}"
    //                                     style="width: 2rem"
    //                                 >
    //                             </div>
    //             `
    //         })
    //         document.getElementById("addShelfToCrawl").classList.remove("skeleton")
    //     }, (err) => {
    //         console.error(err)
    //         document.getElementById("addShelfToCrawl").classList.remove("skeleton")
    //     })
    // }, (err) => {
    //     document.getElementById("addShelfToCrawl").classList.remove("skeleton")
    //     document.getElementById("newShelfToCheckShelfPreviewRow").style.display = "flex"
    //     document.getElementById("shelfUrlAutomatedCheckStatsTextBox").innerHTML +=
    //     `
    //         <div class="col text-center">
    //             <p class="pl-4 mb-0">That's not a valid Goodreads shelf URL. This is an example of a valid shelf URL:</p>
    //             <p class=" pl-4 pb-0"> <a href="https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read">https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read </a> </p>
    //         </div>
    //     `
    // })
})

document.getElementById("settingsTestWebhookURLButton").addEventListener("click", (ev) => {
    const webhookURL = document.getElementById("discordWebhookURLInputBox").value
    testDiscordWebhookURL(webhookURL)
})

document.getElementById("settingsTestOwnedBooksShelfURLButton").addEventListener("click", (ev) => {
    document.getElementById("ownedShelfPreviewRow").style.display = "none"
    document.getElementById("ownedShelfURLCheckStatsBox").innerHTML = ""
    document.getElementById("shelfUrlOwnedStatsTextBox").textContent = ""
    const shelfUrl = document.getElementById("ownedBookshelfURLInputBox").value

    document.getElementById("settingsTestOwnedBooksShelfURLButton").classList.add("skeleton")
    
    setOwnedBooksShelfURL(shelfUrl).then((res) => {
        getPreviewForBookShelf(shelfUrl).then(bookPreview => {
            console.log(bookPreview)
            document.getElementById("ownedShelfPreviewRow").style.display = "flex"
            document.getElementById("ownedShelfURLCheckStatsBox").innerHTML = getHTMLForShelfToCrawl(bookPreview)
            document.getElementById("settingsTestOwnedBooksShelfURLButton").classList.remove("skeleton")
        }, (err) => {
            console.error(err)
            document.getElementById("settingsTestOwnedBooksShelfURLButton").classList.remove("skeleton")
        })
    }, (err) => {
        document.getElementById("settingsTestOwnedBooksShelfURLButton").classList.remove("skeleton")
        document.getElementById("ownedShelfPreviewRow").style.display = "flex"
        document.getElementById("ownedShelfURLCheckStatsBox").innerHTML +=
        `
            <div class="col text-center">
                <p class="pl-4 mb-0">That's not a valid Goodreads shelf URL. This is an example of a valid shelf URL:</p>
                <p class=" pl-4 pb-0"> <a href="https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read">https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read </a> </p>
            </div>
        `
    })
})

function getAndRenderOwnedBooksShelfURL() {
    getOwnedBooksShelfURL().then(url => {
        document.getElementById("ownedBookshelfURLInputBox").value = url
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderDiscordWebhookURL() {
    getDiscordWebhookURL().then(url => {
        document.getElementById("discordWebhookURLInputBox").value = url
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderDiscordMessageFormatPreference() {
    getDiscordMessageFormat().then(format => {
        if (format == "big") {
            highlightBigStyleMessagePreference()
        } else if (format == "small") {
            highlightSmallStyleMessagePreference()
        }
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderAutomatedCrawlTime() {
    getAutomatedCrawlTime().then(time => {
        document.getElementById("automatedCheckTime").textContent = time
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderSendAlertsWhenBookNoLongerAvailable() {
    getSendAlertWhenBookNoLongerAvailable().then(enabled => {
        console.log(enabled)
        if (enabled == true) {
            document.getElementById("sendWebhookWhenNoLongerAvailable").checked = true
        } else {
            document.getElementById("sendWebhookWhenNoLongerAvailable").checked = false
        }
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderSendAlertOnlyWhenFreeShippingKicksIn() {
    getSendAlertOnlyWhenFreeShippingKicksIn().then(enabled => {
        if (enabled == true) {
            document.getElementById("sendWebhookOnlyWhenFreeShippingKicksIn").checked = true
        } else {
            document.getElementById("sendWebhookOnlyWhenFreeShippingKicksIn").checked = false
        }
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderOnlyConsiderEnglishBooks() {
    getOnlyConsiderEnglishBooks().then(enabled => {
        if (enabled == true) {
            document.getElementById("onlyEnglishBooksEnable").checked = true
        } else {
            document.getElementById("onlyEnglishBooksEnable").checked = false
        }
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderKnownAuthorsList() {
    document.getElementById("knownAuthors").innerHTML = ""
    getKnownAuthorList().then(authors => {
        const ignoredAuthors = authors.filter(author => {
            return author.ignore === true
        })
        const nonIgnoredAuthors = authors.filter(author => {
            return author.ignore === false
        })
        ignoredAuthors.forEach(ignoredAuthor => {
            document.getElementById("knownAuthors").innerHTML += 
            `
            
            <div class="pl-2 pr-2 pt-1 pb-1 ml-1 mr-1 mt-2 text-center recentCrawlBox ignoreAuthor" style="${swayaaangBorders(0.5)} border: 2px solid #a33131; font-size: 0.8rem; background-color: #a33131" id="${ignoredAuthor.name}"> 
                ${ignoredAuthor.name}
            </div>
            `
        })

        nonIgnoredAuthors.forEach(author => {
            document.getElementById("knownAuthors").innerHTML += 
            `
            <div class="pl-2 pr-2 pt-1 pb-1 ml-1 mr-1 mt-2 text-center recentCrawlBox ignoreAuthor" style="${swayaaangBorders(0.5)} border: 2px solid #c0c0c0; font-size: 0.8rem;" id="${author.name}"> 
                ${author.name}
            </div>
            `
        })

        document.querySelectorAll(".ignoreAuthor").forEach(element => {
            element.addEventListener("click", (ev) => {
                console.log("click")
                toggleAuthorIgnore(ev.target.id).then(() => {
                    getAndRenderKnownAuthorsList()
                }, err => {
                    console.error(err)
                })
            })
        })
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderShelvesToCrawl() {
    document.getElementById("newShelfToCheckShelfPreviewRow").style.display = "none"
    document.getElementById("automatedCheckShelfURLCheckStatsBox").innerHTML = ""
    document.getElementById("shelfUrlAutomatedCheckStatsTextBox").textContent = ""

    document.getElementById("shelvesToCrawlRow").innerHTML = ""
    getShelvesToCrawl().then(shelvesToCrawl => {
        console.log(shelvesToCrawl)
        shelvesToCrawl.forEach(shelfToCrawl => {
            document.getElementById("shelvesToCrawlRow").innerHTML += getHTMLForShelfToCrawl(shelfToCrawl)
        })

        document.querySelectorAll(".shelfToCrawlElem").forEach(element => {
            element.addEventListener("click", (ev) => {
                console.log(`clicl ${ev.target.id}`)
                removeShelfToCrawl(ev.target.id).then((res) => {
                    getAndRenderShelvesToCrawl()
                }, err => {
                    console.error(err)
                })
            })
        })
    })
}

function getHTMLForShelfToCrawl(shelfToCrawl) {
    return `
    <div class="m-1 pb-2 thinBorder" style="width: 47%;">
    <div class="col">
        <div class="row">
            <div class="col-11">

            </div>
            <div class="col pl-0 pr-0 text-center" style="font-size: 0.7rem">
                <a href="#" class="shelfToCrawlElem" id="${shelfToCrawl.shelfURL}">
                    x
                </a>
            </div>
        </div>

        <div class="row">
            <div class="col-7 pr-1 bookCoverPreviewsCol" style="${shelfToCrawl.coversPreview.length > 8 ? "overflow-x: scroll; cursor: grab; white-space: nowrap;" : ""} width: 100%; height: 4.5rem">
                ${getAndRenderBookCoverPreviews(shelfToCrawl.coversPreview)}
            </div>
            <div class="col">
                <div class="row" style="height: 70%;">
                    <div class="col text-center">
                        <p class="bookPreviewCrawlKeyTitle">
                            <a href="${shelfToCrawl.shelfURL}">${shelfToCrawl.crawlKey}</a>
                        </p>
                    </div>
                </div>
                <div class="row">
                    <div class="col text-center bookPreviewBookCount">
                        ${shelfToCrawl.bookCount} ${shelfToCrawl.bookCount > 1 ? "books" : "book"}
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

    `
}

function getAndRenderBookCoverPreviews(covers) {
    let output = ``
    let currLeftPercentage = 5

    covers.forEach(cover => {
        output +=
        `
        <img
            src="${cover}"
            style="position: absolute; left:${currLeftPercentage}%; height: 4rem"
        >
        `
        currLeftPercentage += 10
    })

    return output
}

document.getElementById("sendWebhookOnlyWhenFreeShippingKicksIn").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setSendAlertOnlyWhenFreeShippingKicksIn("true")
    } else {
        setSendAlertOnlyWhenFreeShippingKicksIn("false")
    }
})

document.getElementById("onlyEnglishBooksEnable").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setOnlyConsiderEnglishBooks("true")
    } else {
        setOnlyConsiderEnglishBooks("false")
    }
})

document.getElementById("addMoreAuthorBooksToAvailableList").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setAddMoreBooksFromAuthorToAvailableBooksList("true")
    } else {
        setAddMoreBooksFromAuthorToAvailableBooksList("false")
    }
})

document.getElementById("setsearchotherseriesbookslookup").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setSearchOtherSeriesBooksLookup("true")
    } else {
        setSearchOtherSeriesBooksLookup("false")
    }
})


function getAndRenderAddMoreAuthorBooksToAvailableList() {
    getAddMoreBooksFromAuthorToAvailableBooksList().then(enabled => {
        if (enabled == true) {
            document.getElementById("addMoreAuthorBooksToAvailableList").checked = true
        } else {
            document.getElementById("addMoreAuthorBooksToAvailableList").checked = false
        }
    }, (err) => {
        console.error(err)
    })
}

function getAndRenderSeriesInAutomatedCrawl() {
    getsearchotherseriesbooksinlookup().then(enabled => {
        if (enabled == true) {
            document.getElementById("setsearchotherseriesbookslookup").checked = true
        } else {
            document.getElementById("setsearchotherseriesbookslookup").checked = false
        }
    }, (err) => {
        console.error(err)
    })
}

document.getElementById("sendWebhookWhenNoLongerAvailable").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setSendAlertWhenBookNoLongerAvailable("true")
    } else {
        setSendAlertWhenBookNoLongerAvailable("false")
    }
})

document.getElementById("purgeIgnoredAuthorsButton").addEventListener("click", (ev) => {
    purgeIgnoredAuthorsFromAvailableBooks()
})

document.getElementById("clearKnownAuthors").addEventListener("click", (ev) => {
    clearKnownAuthors().then(() => {
        getAndRenderKnownAuthorsList()
    }, (err) => {
        console.error(err)
    })

})

document.getElementById("disableAutomatedChecks").addEventListener("click", (ev) => {
    disableAutomatedChecks().then(() => {
        getAndRenderAutomatedCrawlTime()
    }, (err) => {
        console.error(err)
    })
})

document.getElementById("disableDiscordNotifications").addEventListener("click", (ev) => {
    clearDiscordWebhook("").then(() => {
        getAndRenderDiscordWebhookURL()
    }, (err) => {
        console.error(err)
    })
})

document.getElementById("purgeAuthorMatches").addEventListener("click", (ev) => {
    purgeAuthorMatches().then(() => {
        
    }, (err) => {
        console.error(err)
    })
})

document.getElementById("purgeSeriesMatches").addEventListener("click", (ev) => {
    purgeSeriesMatches().then(() => {
        
    }, (err) => {
        console.error(err)
    })
})

function giveSwayaaangBordersToItems() {
    document.getElementById("availableLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("shelfLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("settingsLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("seriesLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("bigStyleBox").style = swayaaangBorders(1.6)
    document.getElementById("smallStyleBox").style = swayaaangBorders(1.6)
    document.getElementById("addShelfToCrawl").style = swayaaangBorders(0.4)
    document.getElementById("settingsTestWebhookURLButton").style = swayaaangBorders(0.4)
    document.getElementById("settingsTestOwnedBooksShelfURLButton").style = swayaaangBorders(0.4)
    document.getElementById("settingsSetAutomatedCheckTimeButton").style = swayaaangBorders(0.4)
    document.getElementById("purgeIgnoredAuthorsButton").style = swayaaangBorders(0.6)
    document.getElementById("clearKnownAuthors").style = swayaaangBorders(0.6)
    document.getElementById("disableAutomatedChecks").style = swayaaangBorders(0.6)
    document.getElementById("disableDiscordNotifications").style = swayaaangBorders(0.6)
    document.getElementById("purgeAuthorMatches").style = swayaaangBorders(0.6)
    document.getElementById("purgeSeriesMatches").style = swayaaangBorders(0.6)
}

document.getElementById("bigStyleBox").addEventListener("click", () => {
    highlightBigStyleMessagePreference()
    setDiscordMessageFormat("big")
})

document.getElementById("smallStyleBox").addEventListener("click", () => {
    highlightSmallStyleMessagePreference()
    setDiscordMessageFormat("small")
})

function highlightBigStyleMessagePreference() {
    document.getElementById("bigStyleBox").classList.add("selectedBackground")
    document.getElementById("smallStyleBox").classList.remove("selectedBackground")
}

function highlightSmallStyleMessagePreference() {
    document.getElementById("smallStyleBox").classList.add("selectedBackground")
    document.getElementById("bigStyleBox").classList.remove("selectedBackground")
}

function getSWAGEstimateForCrawlTime(bookCount) {
    return bookCount * 1.3
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

function clearList() {
    return new Promise((resolve, reject) => {
        fetch(`/resetavailablebooks`, {
            method: 'POST',
        })
        .then((res) => {
            resolve(res)
        }, (err) => {
            reject(err)
        });
    })
}

function getOwnedBooksShelfURL(){
    return new Promise((resolve, reject) => {
        fetch(`/settings/getownedbooksshelfurl`)
        .then((res) => res.json())
        .then((res) => {
            resolve(res.shelfURL)
        }, (err) => {
            reject(err)
        });
    })
}

function setOwnedBooksShelfURL(shelfURL){
    return new Promise((resolve, reject) => {
        fetch(`/settings/setownedbooksshelfurl?shelfurl=${encodeURIComponent(shelfURL)}`, {
            method: "POST"
        })
        .then((res) => res.json())
        .then((res) => {
            if (res.hasOwnProperty("error")) {
                reject()
            } else {
                resolve(res)
            }
        }, (err) => {
            reject(err)
        });
    })
}

function testDiscordWebhookURL(webhookURL) {
    fetch(`/settings/testdiscordwebhook?webhookurl=${webhookURL}`, {
        method: "GET",
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json"
        },
    })
}

function getAutomatedCrawlTime() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getautomatedcrawltime`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.time)
        }, (err) => {
            reject(err)
        });
    })
}

function setAutomatedCrawlTime(time) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setautomatedcrawltime?time=${encodeURIComponent(time)}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function disableAutomatedChecks(time) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/disableautomatedcrawltime`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function setDiscordWebhookURL(webhookURL) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setdiscordwebhook?webhookurl=${webhookURL}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function getDiscordWebhookURL(webhookURL) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getdiscordwebhook`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.webhook)
        }, (err) => {
            reject(err)
        });
    })
}

function clearDiscordWebhook(webhookURL) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/cleardiscordwebhook`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve(res.webhook)
        }, (err) => {
            reject(err)
        });
    })
}

function purgeAuthorMatches() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/purgeauthormatches`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then(() => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function purgeSeriesMatches() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/purgeseriesmatches`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then(() => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}


function setDiscordMessageFormat(messageFormat) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setdiscordmessageformat?messageformat=${messageFormat}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function getDiscordMessageFormat(webhookURL) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getdiscordmessageformat`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.format)
        }, (err) => {
            reject(err)
        });
    })
}

function setSendAlertWhenBookNoLongerAvailable(enabled) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setsendalertwhenbooknolongeravailable?enabled=${enabled}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function getSendAlertWhenBookNoLongerAvailable() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getsendalertwhenbooknolongeravailable`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.enabled)
        }, (err) => {
            reject(err)
        });
    })
}

function setSendAlertOnlyWhenFreeShippingKicksIn(enabled) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setsendalertonlywhenfreeshippingkicksin?enabled=${enabled}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function getSendAlertOnlyWhenFreeShippingKicksIn() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getsendalertonlywhenfreeshippingkicksin`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.enabled)
        }, (err) => {
            reject(err)
        });
    })
}

function getOnlyConsiderEnglishBooks() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getonlyenglishbooksenabled`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.enabled)
        }, (err) => {
            reject(err)
        });
    })
}

function setOnlyConsiderEnglishBooks(enabled) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setonlyenglishbooksenabled?enabled=${enabled}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}


function setAddMoreBooksFromAuthorToAvailableBooksList(enabled) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setaddmoreauthorbookstoavailablelist?enabled=${enabled}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function getAddMoreBooksFromAuthorToAvailableBooksList() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getaddmoreauthorbookstoavailablelist`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.enabled)
        }, (err) => {
            reject(err)
        });
    })
}

function setSearchOtherSeriesBooksLookup(enabled) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/setsearchotherseriesbookslookup?enabled=${enabled}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function getsearchotherseriesbooksinlookup() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getsearchotherseriesbooksinlookup`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.enabled)
        }, (err) => {
            reject(err)
        });
    })
}

function getPreviewForBookShelf(shelfURL) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getpreviewforshelf?shelfurl=${encodeURIComponent(shelfURL)}`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.shelfToCrawlPreview)
        }, (err) => {
            reject(err)
        });
    })
}

function getKnownAuthorList() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getknownauthors`, {
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

function clearKnownAuthors() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/clearknownauthors`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function toggleAuthorIgnore(author, enable) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/toggleauthorignore?author=${encodeURIComponent(author)}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function purgeIgnoredAuthorsFromAvailableBooks(author) {
    return new Promise((resolve, reject) => {
        fetch(`/purgeignoredauthorsfromavailablebooks`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then(() => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function getShelvesToCrawl() {
    return new Promise((resolve, reject) => {
        fetch(`/settings/getshelvestocrawl`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            resolve(res.shelvesToCrawlPreviews)
        }, (err) => {
            reject(err)
        });
    })
}

function addShelfToCrawl(shelfURL) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/addshelftocrawl?shelfurl=${encodeURIComponent(shelfURL)}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            if (res.status != 200) {
                reject(res.error)
            }
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function removeShelfToCrawl(shelfURL) {
    return new Promise((resolve, reject) => {
        fetch(`/settings/removeshelftocrawl?shelfurl=${encodeURIComponent(shelfURL)}`, {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}