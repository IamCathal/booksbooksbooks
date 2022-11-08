# booksbooksbooks

I use GoodReads for my book reccomendations and I buy my books mostly from TheBookshop.ie. I made this project to be able to automatically check if books on my "shopping list" are available on TheBookshop. Still in active development but hoping to make a 1.0 release before the end of November 2022.

|      |  |
| ----------- | ----------- |
| ![](https://i.imgur.com/TEFxUnN.png)     | ![](https://i.imgur.com/vzhiiJ1.png)   |


### Usage

* `docker-compose up`
* Go to [http://localhost:2945/settings](http://localhost:2945/settings) to set the automated crawl shelf URL, time and discord webhook URL (if you want alerts)
* Start a manual crawl on the homepage or wait for an automated crawl to start on its own (if you've set a shelf URL and time for it to use)

From the homepage new shelves can be manually crawled. From [http://localhost:2945/available](http://localhost:2945/available) all books which have been found so far on TheBookshop.ie will be aggregrated. A total cost will be displayed out of €20 because they offer free shipping on orders above that amount. 

For automated checking you first need to set the (goodreahds) shelf to crawl which can be done onthe [http://localhost:2945/settingsf](http://localhost:2945/settings) page. This will then use your specified shelf and crawl it in the backgrund when the time comes

### Setting Flags

A good few things can be customised from the [http://localhost:2945/settings](http://localhost:2945/settings) page

- [x] Set the time when the automated crawl should be executed. This crawl happens once per day at the specified time
- [x] Send an alert when a book was marked as available from the last crawl but is now no longer available (it was bought)
- [x] Send an alert only when the total cost of the available books has exceeded €20 which means the order is eligible for free shipping
- [x] Whether to have a compact or spacious alert styled messaged

## Security

Unsanitised user input is written straight to the redis instance and although its not entirely sensitive your supplied discord webhook is accessable through the settings page. Do not publically host this service. I'm currently running this on a cloud VPS (that has all ports blocked) and I can access it on my home network through a [tailscale](https://tailscale.com/) network.

## Testing

Have the local instance of redis running on port 6379 and away you go. Why hardcode in the redis to be local? Because its a test and it doesn't matter, its always going to be local. Use `go test -v ./...` in the base directory to run everything. I prefer using [gotestsum](https://github.com/gotestyourself/gotestsum) since its a bit nicer and I use this alias `alias gotall='gotestsum --format=testname -- -v ./...'` to run all tests and `alias gotestcov='go test -v -p 1 ./... -cover -coverprofile=coverage.out --tags=service && go tool cover -html=coverage.out -o coverage.html && firefox coverage.html'` to get coverage and open the output in firefox