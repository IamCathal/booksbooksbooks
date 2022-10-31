let availableBooks = []

giveSwayaaangBordersToItems()
getAndRenderAutomatedShelfCheckURL()

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

function getAndRenderAutomatedShelfCheckURL() {
    getAutomatedShelfCheckURL().then(url => {
        console.log(url)
        document.getElementById("automatedCheckShelfInput").value = url
    }, (err) => {
        console.error(err)
    })
}

function renderAvailableBooks() {
    document.getElementById("availableBooks").innerHTML = ""
    let totalBookCost = 0
    availableBooks.forEach(book => {
        totalBookCost += getBookCost(book.bookPurchaseInfo.price)
        console.log(book)
        document.getElementById("availableBooks").innerHTML +=
        `
                    <div class="col-3 ml-3 pt-3 searchResultBook">
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
                            </div>
                        </div>
                    </div>
        `
    })

    document.getElementById("priceStatsDiv").textContent = `€${totalBookCost.toFixed(2)} / €20`
}

function getBookCost(bookCostString) {
    return parseFloat(bookCostString.replace("€",""))
}


function getAvailableBooks() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2945/getavailablebooks`)
        .then((res) => res.json())
        .then((res) => {
            resolve(res)
        }, (err) => {
            reject(err)
        });
    })
}

function getAutomatedShelfCheckURL(){
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2945/getautomatedbookshelfcheckurl`)
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
        fetch(`http://localhost:2945/setautomatedbookshelfcheckurl?shelfurl=${encodeURIComponent(shelfURL)}`)
        .then((res) => {
            resolve(res)
        }, (err) => {
            reject(err)
        });
    })
}

document.getElementById("automatedCheckShelfInput").addEventListener("keyup", function(event) {
    const shelfUrl = document.getElementById("automatedCheckShelfInput").value
    if (event.key === "Enter") {
        setAutomatedShelfCheckURL(shelfUrl).then((res) => {
            console.log(res)
        }, (err) => {
            console.error(err)
        })
    }
});

function giveSwayaaangBordersToItems() {
    document.getElementById("availableLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("shelfLinkBox").style = swayaaangBorders(0.8)

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
        fetch(`http://localhost:2945/resetavailablebooks`, {
            method: 'POST',
        })
        .then((res) => {
            resolve(res)
        }, (err) => {
            reject(err)
        });
    })
}