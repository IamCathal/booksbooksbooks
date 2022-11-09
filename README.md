# booksbooksbooks

I use GoodReads for my book reccomendations and I buy my books mostly from TheBookshop.ie. I made this project to be able to automatically check if books on my "shopping list" are available on TheBookshop. Still in active development but hoping to make a 1.0 release before the end of November 2022.

|      |  |
| ----------- | ----------- |
| ![](https://i.imgur.com/TEFxUnN.png)     | ![](https://i.imgur.com/vzhiiJ1.png)   |


### Usage

* `docker-compose up`
* Go to [http://localhost:2945/settings](http://localhost:2945/settings)
    * **Automated Check Shelf URL:** URL of the goodreads shelf that will crawled when the automated check is triggered
    * **Automated Check Time:** The time that the automated crawl will be triggered on a daily basis
    * **Discord Webhook URL:** For alerts paste in a discord webhook URL here. See [here on how to go about doing that](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks)
* Automated crawls are now set up. If you want to manually crawl a shelf go to [http://localhost:2945/shelf](http://localhost:2945/shelf)
* Once a few books have been found (through individual crawls with any supplied shelves or through automated crawls) they'll appear on the [http://localhost:2945/available](http://localhost:2945/available) page. Books can be ignored if you don't plan on buying them and unignored if you later change your mind

### Setting Flags

A good few levers can be pulled from the [http://localhost:2945/settings](http://localhost:2945/settings) page to customise your experience

- [x] Set the time when the automated crawl should be executed. The automated crawl happens once per day at the specified time
- [x] Send an alert when a book was marked as available from the last crawl but is now no longer available (it was bought)
- [x] Send an alert only when the total cost of the available books has exceeded €20 which means the order is eligible for free shipping
- [x] Whether to have a compact or spacious alert styled messages in discord

## Security

Unsanitised user input is written straight to the redis instance, unsanitised user input is rendered as raw HTML and although its not entirely sensitive your supplied discord webhook is accessable through the settings page. Do not publically host this service. I'm currently running this on a cloud VPS (that has all ports blocked) and I can access it on my home network through a [tailscale](https://tailscale.com/) setup.

## Testing

Have the local instance of redis running on port 6379 and away you go. Why hardcode in the redis to be local? Because its a test and it doesn't matter, its always going to be local. Use `go test -v ./...` in the base directory to run everything. I prefer using [gotestsum](https://github.com/gotestyourself/gotestsum) since its a bit nicer and I use this alias `alias gotall='gotestsum --format=testname -- -v ./...'` to run all tests and `alias gotestcov='go test -v -p 1 ./... -cover -coverprofile=coverage.out --tags=service && go tool cover -html=coverage.out -o coverage.html && firefox coverage.html'` to get coverage and open the output in firefox