let availableBooks = []

getAvailableBooks().then((res) => {
    availableBooks = res
    renderAvailableBooks(res)
}, err => {
    console.error(err)
})

function renderAvailableBooks() {
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