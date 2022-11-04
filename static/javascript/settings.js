
giveSwayaaangBordersToItems()

getAndRenderSettings()

function getAndRenderSettings() {
    getAndRenderAutomatedShelfCheckURL()
    getAndRenderDiscordWebhookURL()
    getAndRenderDiscordMessageFormatPreference()
    getAndRenderAutomatedCrawlTime()
    getAndRenderSendAlertsWhenBookNoLongerAvailable()
    getAndRenderSendAlertOnlyWhenFreeShippingKicksIn()
}

document.getElementById("settingsSetAutomatedCheckTimeButton").addEventListener("click", (ev) => {
    Timepicker.showPicker({
        onSubmit: (time) => {
            document.getElementById("automatedCheckTime").textContent = time.formatted()
            setAutomatedCrawlTime(time.formatted())
        }
    })
})

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
        if (enabled == "true") {
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
        if (enabled == "true") {
            document.getElementById("sendWebhookOnlyWhenFreeShippingKicksIn").checked = true
        } else {
            document.getElementById("sendWebhookOnlyWhenFreeShippingKicksIn").checked = false
        }
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




document.getElementById("sendWebhookWhenNoLongerAvailable").addEventListener("change", (ev) => {
    if (ev.currentTarget.checked) {
        setSendAlertWhenBookNoLongerAvailable("true")
    } else {
        setSendAlertWhenBookNoLongerAvailable("false")
    }
})

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
        fetch(`/settings/setautomatedbookshelfcheckurl?shelfurl=${encodeURIComponent(shelfURL)}`)
        .then((res) => {
            resolve(res)
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