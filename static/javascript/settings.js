giveSwayaaangBordersToItems()
getAndRenderAutomatedShelfCheckURL()

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

document.getElementById("shelfCheckURLInputBox").addEventListener("keyup", function(event) {
    const shelfUrl = document.getElementById("shelfCheckURLInputBox").value
    if (event.key === "Enter") {
        setAutomatedShelfCheckURL(shelfUrl).then((res) => {
            console.log(res)
        }, (err) => {
            console.error(err)
        })
    }
});

document.getElementById("settingsTestWebhookURLButton").addEventListener("click", (ev) => {
    const webhookURL = document.getElementById("discordWebhookURLInputBox").value
    testDiscordWebhookURL(webhookURL)
})

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

function getAndRenderAutomatedShelfCheckURL() {
    getAutomatedShelfCheckURL().then(url => {
        console.log(url)
        document.getElementById("shelfCheckURLInputBox").value = url
    }, (err) => {
        console.error(err)
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

function testDiscordWebhookURL(webhookURL) {
    fetch(`http://localhost:2945/testdiscordwebhook?webhookurl=${webhookURL}`, {
        method: "GET",
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json"
        },
    })
}

function setDiscordWebhookURL(webhookURL) {
    fetch(`http://localhost:2945/setdiscordwebhook?webhookurl=${webhookURL}`, {
        method: "GET",
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json"
        },
    })
}

function getDiscordWebhookURL(webhookURL) {
    fetch(`http://localhost:2945/getdiscordwebhook?webhookurl=${webhookURL}`, {
        method: "GET",
        headers: {
            "Content-Type": "application/json",
            "Accept": "application/json"
        },
    })
}