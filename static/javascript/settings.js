let ignoredAuthorsList = []

giveSwayaaangBordersToItems()

getAndRenderSettings()

function getAndRenderSettings() {
    getAndRenderAutomatedShelfCheckURL()
    getAndRenderDiscordWebhookURL()
    getAndRenderDiscordMessageFormatPreference()
    getAndRenderAutomatedCrawlTime()
    getAndRenderSendAlertsWhenBookNoLongerAvailable()
    getAndRenderSendAlertOnlyWhenFreeShippingKicksIn()
    getAndRenderAddMoreAuthorBooksToAvailableList()
    getAndRenderKnownAuthorsList()
}

document.getElementById("settingsSetAutomatedCheckTimeButton").addEventListener("click", (ev) => {
    Timepicker.showPicker({
        onSubmit: (time) => {
            document.getElementById("automatedCheckTime").textContent = time.formatted()
            setAutomatedCrawlTime(time.formatted())
        }
    })
})

document.getElementById("settingsTestShelfURLButton").addEventListener("click", (ev) => {
    document.getElementById("shelfPreviewRow").style.display = "none"
    document.getElementById("shelfURLCheckStatsBox").innerHTML = ""
    document.getElementById("shelfUrlCheckStatsTextBox").textContent = ""
    const shelfUrl = document.getElementById("shelfCheckURLInputBox").value

    document.getElementById("settingsTestShelfURLButton").classList.add("skeleton")
    
    setAutomatedShelfCheckURL(shelfUrl).then((res) => {
        getPreviewForBookShelf(shelfUrl).then(bookPreview => {
            document.getElementById("shelfPreviewRow").style.display = "flex"
    
            document.getElementById("shelfUrlCheckStatsTextBox").textContent = 
                `Found ${bookPreview.totalBooks} books. Should take roughly ${Math.floor(getSWAGEstimateForCrawlTime(bookPreview.totalBooks))}s to crawl`
            bookPreview.books.forEach(book => {
                document.getElementById("shelfURLCheckStatsBox").innerHTML += `
                                <div class="col-1 pr-1">
                                    <img 
                                        src="${book.cover}"
                                        style="width: 2.5rem"
                                    >
                                </div>
                `
            })
            document.getElementById("settingsTestShelfURLButton").classList.remove("skeleton")
        }, (err) => {
            console.error(err)
            document.getElementById("settingsTestShelfURLButton").classList.remove("skeleton")
        })
    }, (err) => {
        document.getElementById("settingsTestShelfURLButton").classList.remove("skeleton")
        document.getElementById("shelfPreviewRow").style.display = "flex"
        document.getElementById("shelfURLCheckStatsBox").innerHTML +=
        `
            <div class="col text-center">
                <p class="pl-4 mb-0">That's not a valid Goodreads shelf URL. This is an example of a valid shelf URL:</p>
                <p class=" pl-4 pb-0"> <a href="https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read">https://www.goodreads.com/review/list/26367680-stephen-king?shelf=read </a> </p>
            </div>
        `
    })
})

document.getElementById("settingsTestWebhookURLButton").addEventListener("click", (ev) => {
    const webhookURL = document.getElementById("discordWebhookURLInputBox").value
    testDiscordWebhookURL(webhookURL)
})

function getAndRenderAutomatedShelfCheckURL() {
    getAutomatedShelfCheckURL().then(url => {
        document.getElementById("shelfCheckURLInputBox").value = url
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

function getAndRenderKnownAuthorsList() {
    document.getElementById("knownAuthors").innerHTML = ""
    getKnownAuthorList().then(authors => {
        const ignoredAuthors = authors.filter(author => {
            return author.ignore === true
        })
        ignoredAuthorsList = ignoredAuthors
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

document.getElementById("sendWebhookOnlyWhenFreeShippingKicksIn").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setSendAlertOnlyWhenFreeShippingKicksIn("true")
    } else {
        setSendAlertOnlyWhenFreeShippingKicksIn("false")
    }
})

document.getElementById("addMoreAuthorBooksToAvailableList").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setAddMoreBooksFromAuthorToAvailableBooksList("true")
    } else {
        setAddMoreBooksFromAuthorToAvailableBooksList("false")
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

document.getElementById("sendWebhookWhenNoLongerAvailable").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setSendAlertWhenBookNoLongerAvailable("true")
    } else {
        setSendAlertWhenBookNoLongerAvailable("false")
    }
})

document.getElementById("purgeIgnoredAuthorsButton").addEventListener("click", (ev) => {
    ignoredAuthorsList.forEach(author => {
        purgeAuthorFromAvailableBooks(author.name)
    })
})

function giveSwayaaangBordersToItems() {
    document.getElementById("availableLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("shelfLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("settingsLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("bigStyleBox").style = swayaaangBorders(1.6)
    document.getElementById("smallStyleBox").style = swayaaangBorders(1.6)
    document.getElementById("settingsTestShelfURLButton").style = swayaaangBorders(0.4)
    document.getElementById("settingsTestWebhookURLButton").style = swayaaangBorders(0.4)
    document.getElementById("settingsSetAutomatedCheckTimeButton").style = swayaaangBorders(0.4)
    document.getElementById("purgeIgnoredAuthorsButton").style = swayaaangBorders(0.6)

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

function getAutomatedShelfCheckURL(){
    return new Promise((resolve, reject) => {
        fetch(`/settings/getautomatedbookshelfcheckurl`)
        .then((res) => res.json())
        .then((res) => {
            resolve(res.shelfURL)
        }, (err) => {
            reject(err)
        });
    })
}

function setAutomatedShelfCheckURL(shelfURL){
    return new Promise((resolve, reject) => {
        fetch(`/settings/setautomatedbookshelfcheckurl?shelfurl=${encodeURIComponent(shelfURL)}`, {
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
            resolve(res)
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

function purgeAuthorFromAvailableBooks(author) {
    return new Promise((resolve, reject) => {
        fetch(`/purgeauthorfromavailablebooks?author=${encodeURIComponent(author)}`, {
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