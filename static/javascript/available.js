let availableBooks = []

giveSwayaaangBordersToItems()

getAvailableBooks().then((res) => {
    availableBooks = res
    renderAvailableBooks(res)
}, err => {
    console.error(err)
})

document.getElementById("clearList").addEventListener("click", () => {
    clearList()
    getAvailableBooks().then((res) => {
        availableBooks = res
        renderAvailableBooks(res)
    }, err => {
        console.error(err)
    })
})

function renderAvailableBooks(availableBookList) {
    document.getElementById("availableBooks").innerHTML = ""
    document.getElementById("ignoreBookRow").innerHTML = ""
    document.getElementById("priceStatsDiv").innerHTML = ""
    document.getElementById("ignoredPriceStatsDiv").innerHTML = ""
    let totalBookCost = 0
    let totalIgnoredBookCost = 0
    availableBookList.forEach(book => {
        if (book.ignore == false) {
            document.getElementById("availableBooks").innerHTML +=
            `
                        <div class="col-3 pt-3 searchResultBook">
                            <div class="row">
                                <div class="col-3 pl-1">
                                    <a href="${book.bookPurchaseInfo.link}">
                                        <img src="${book.bookPurchaseInfo.cover}" style="width: 3.5rem;" title="">
                                    </a>
                                </div>
                                <div class="col">
                                    <div class="row" style="font-weight: bold; font-size: 0.8rem">
                                        ${book.bookPurchaseInfo.title}
                                    </div>
                                    <div class="row" style="font-size: 0.6rem">
                                        ${book.bookPurchaseInfo.author}
                                    </div>
                                    <div class="row" style="font-weight: bold; font-size: 0.7rem">
                                        ${book.bookPurchaseInfo.price}
                                    </div>
                                    <div class="row ignoreBook" style="font-size: 0.6rem; color: #c0c0c0" id="${book.bookPurchaseInfo.link}">
                                        Ignore this book
                                    </div>
                                </div>
                            </div>
                        </div>
            `
            totalBookCost += getBookCost(book.bookPurchaseInfo.price)
            document.getElementById("priceStatsDiv").textContent = `€${totalBookCost.toFixed(2)} / €20`
        } else {
            document.getElementById("ignoreBookRow").innerHTML += 
            `
                            <div class="col-3 pt-3 searchResultBook text-left">
                                <div class="row">
                                    <div class="col-3 pl-1">
                                        <a href="${book.bookPurchaseInfo.link}">
                                            <img src="${book.bookPurchaseInfo.cover}" style="width: 3.5rem;" title="">
                                        </a>
                                    </div>
                                    <div class="col">
                                        <div class="row" style="font-weight: bold; font-size: 0.8rem">
                                            ${book.bookPurchaseInfo.title}
                                        </div>
                                        <div class="row" style="font-size: 0.6rem">
                                            ${book.bookPurchaseInfo.author}
                                        </div>
                                        <div class="row" style="font-weight: bold; font-size: 0.7rem">
                                            ${book.bookPurchaseInfo.price}
                                        </div>
                                        <div class="row unignoreBook" style="font-size: 0.6rem; color: #c0c0c0" id="${book.bookPurchaseInfo.link}">
                                            Unignore this book
                                        </div>
                                    </div>
                                </div>
                            </div>
            `
            totalIgnoredBookCost += getBookCost(book.bookPurchaseInfo.price)
            document.getElementById("ignoredPriceStatsDiv").textContent = `€${totalIgnoredBookCost.toFixed(2)} / €20`
        }

        document.querySelectorAll(".unignoreBook").forEach(element => {
            element.addEventListener("click", (ev) => {
                unignoreBook(ev.target.id).then((res) => {
                    getAvailableBooks().then(newAvailableBooks => {
                        renderAvailableBooks(newAvailableBooks)
                    }, err => {
                        console.error(err)
                    })
                }, err => {
                    console.error(err)
                })
            })
        })

        document.querySelectorAll(".ignoreBook").forEach(element => {
            element.addEventListener("click", (ev) => {
                ignoreBook(ev.target.id).then((res) => {
                    getAvailableBooks().then(newAvailableBooks => {
                        renderAvailableBooks(newAvailableBooks)
                    }, err => {
                        console.error(err)
                    })
                }, err => {
                    console.error(err)
                })
            })
        })
    })
}

function getBookCost(bookCostString) {
    return parseFloat(bookCostString.replace("€",""))
}


function getAvailableBooks() {
    return new Promise((resolve, reject) => {
        fetch(`/getavailablebooks`)
        .then((res) => res.json())
        .then((res) => {
            resolve(res)
        }, (err) => {
            reject(err)
        });
    })
}

function giveSwayaaangBordersToItems() {
    document.getElementById("availableLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("shelfLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("settingsLinkBox").style = swayaaangBorders(0.8)
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

function ignoreBook(bookURL) {
    return new Promise((resolve, reject) => {
        fetch(`/ignorebook?bookurl=${encodeURIComponent(bookURL)}`, {
            method: 'POST',
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}

function unignoreBook(bookURL) {
    return new Promise((resolve, reject) => {
        fetch(`/unignorebook?bookurl=${encodeURIComponent(bookURL)}`, {
            method: 'POST',
        })
        .then((res) => {
            resolve()
        }, (err) => {
            reject(err)
        });
    })
}