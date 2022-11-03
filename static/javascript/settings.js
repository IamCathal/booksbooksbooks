
giveSwayaaangBordersToItems()

getAndRenderSettings()

function getAndRenderSettings() {
    getAndRenderAutomatedShelfCheckURL()
    getAndRenderDiscordWebhookURL()
    getAndRenderDiscordMessageFormatPreference()
    getAndRenderAutomatedCrawlTime()
}

document.getElementById("settingsSetAutomatedCheckTimeButton").addEventListener("click", (ev) => {
    Timepicker.showPicker({
        onSubmit: (time) => {
            document.getElementById("automatedCheckTime").textContent = time.formatted()
            setAutomatedCrawlTime(time.formatted())
        }
    })
})

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
        console.log(`time was ${time}`)
        document.getElementById("automatedCheckTime").textContent = time
    }, (err) => {
        console.error(err)
    })
}

function giveSwayaaangBordersToItems() {
    document.getElementById("availableLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("shelfLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("settingsLinkBox").style = swayaaangBorders(0.8)
    document.getElementById("bigStyleBox").style = swayaaangBorders(1.6)
    document.getElementById("smallStyleBox").style = swayaaangBorders(1.6)

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

function getAutomatedCrawlTime() {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2945/settings/getautomatedcrawltime`, {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
                "Accept": "application/json"
            },
        }).then((res) => res.json())
        .then((res) => {
            console.log(`got back: ${res}`)
            resolve(res.time)
        }, (err) => {
            reject(err)
        });
    })
}

function setAutomatedCrawlTime(time) {
    return new Promise((resolve, reject) => {
        fetch(`http://localhost:2945/settings/setautomatedcrawltime?time=${encodeURIComponent(time)}`, {
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
        fetch(`http://localhost:2945/setdiscordwebhook?webhookurl=${webhookURL}`, {
            method: "GET",
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
        fetch(`http://localhost:2945/getdiscordwebhook`, {
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
        fetch(`http://localhost:2945/settings/setdiscordmessageformat?messageformat=${messageFormat}`, {
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
        fetch(`http://localhost:2945/settings/getdiscordmessageformat`, {
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