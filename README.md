# booksbooksbooks

I use [GoodReads](https://www.goodreads.com/) for my book reccomendations and I buy my books mostly from [TheBookshop.ie](https://thebookshop.ie/). I made this project to be able to automatically check if books on my "shopping list" are available on TheBookshop. Still in active development and if there is not an official release version then don't expect even master to work properly all of the time (I do use it myself at the moment so it should work).

|      |  |
| ----------- | ----------- |
| ![](https://i.imgur.com/TEFxUnN.png)     | ![](https://i.imgur.com/vzhiiJ1.png)   |


### Usage

* `docker-compose up` and away you go
* Go to [http://localhost:2945/settings](http://localhost:2945/settings) page and fill in all the details you need

### Setting Flags

A good few levers can be pulled from the [http://localhost:2945/settings](http://localhost:2945/settings) page to customise your experience

- [x] If more books are found from an author who's in your shelf then add those books to the available list
- [x] If a book is in a series then lookup all other books in that series
- [x] Use discord webhooks for updates (very handy) about when books become available
- [x] Alert when a book that was previously marked as available is no longer for sale
- [x] Send an alert only when the total cost of the available books has exceeded €20 which means the order is eligible for free shipping
- [x] Filter out specific authors from future search results so their books won't ever show up as available

## Security

Unsanitised user input is written straight to the redis instance and is rendered as raw HTML and although its not entirely sensitive your supplied discord webhook is accessable through the settings page. Do not publically host this service. I'm currently running this on a cloud VPS (that has all ports blocked) and I can access it on my home network through a [tailscale](https://tailscale.com/) setup.

## Testing

Have the local instance of redis running on port 6379 and away you go. Why hardcode in the redis to be local? Because its a test and it doesn't matter, its always going to be local. Use `go test -v ./...` in the base directory to run everything. I prefer using [gotestsum](https://github.com/gotestyourself/gotestsum) since its a bit nicer and I use this alias `alias gotall='gotestsum --format=testname -- -v ./...'` to run all tests and `alias gotestcov='go test -v -p 1 ./... -cover -coverprofile=coverage.out --tags=service && go tool cover -html=coverage.out -o coverage.html && firefox coverage.html'` to get coverage and open the output in firefox