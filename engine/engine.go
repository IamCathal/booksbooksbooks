package engine

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
)

var (
	BOOKS_DISPLAYED_PER_PAGE = 30
)

func Worker(shelfURL string, ws *websocket.Conn) {

	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		panic("invalid shelf url")
	}

	shelfStatsChan := make(chan int, 1)
	booksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchResultsFromTheBookshopChan := make(chan dtos.AllBookshopBooksSearchResults, 200)

	fmt.Printf("Retrieving books from shelf: %s\n", shelfURL)
	goodreads.GetBooksFromShelf(shelfURL, shelfStatsChan, booksFoundFromGoodReadsChan)

	booksFound := 0
	searchResultsReturned := 0
	totalBooksFromGoodReads := -1
	currCrawlStats := dtos.CrawlStats{
		TotalBooks:    totalBooksFromGoodReads,
		BooksCrawled:  booksFound,
		BooksSearched: searchResultsReturned,
	}

	for {
		if allBooksFound(currCrawlStats) {
			break
		}

		select {
		case totalBooks := <-shelfStatsChan:
			currCrawlStats.TotalBooks = totalBooks
			writeTotalBooksMsg(currCrawlStats, ws)
			close(shelfStatsChan)

		case bookFromGoodReads := <-booksFoundFromGoodReadsChan:
			currCrawlStats.BooksCrawled++
			fmt.Printf("(%d) found a book: %+v\n", booksFound, bookFromGoodReads)
			writeGoodReadsBookMsg(bookFromGoodReads, currCrawlStats, ws)
			thebookshop.SearchForBook(bookFromGoodReads, searchResultsFromTheBookshopChan)

		case searchResultFromTheBookshop := <-searchResultsFromTheBookshopChan:
			currCrawlStats.BooksSearched++
			fmt.Printf("\nSearch result found: %+v\n\n", searchResultFromTheBookshop)
			writeSearchResultReturnedMsg(searchResultFromTheBookshop, currCrawlStats, ws)

		}
	}
	fmt.Printf("Exiting. All books queried from Goodreads")
}

func allBooksFound(crawlStats dtos.CrawlStats) bool {
	if ((crawlStats.BooksCrawled == crawlStats.TotalBooks) &&
		(crawlStats.BooksSearched == crawlStats.TotalBooks)) && crawlStats.TotalBooks != -1 {
		return true
	}
	return false
}
