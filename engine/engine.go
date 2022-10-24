package engine

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
)

var (
	BOOKS_DISPLAYED_PER_PAGE = 30
)

func Worker(shelfURL string, ws *websocket.Conn) {
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfURL '%s' given", shelfURL)
		writeErrorMsg(errorMsg, ws)
		return
	}

	db.SaveRecentCrawlStats(shelfURL)

	shelfStatsChan := make(chan int, 1)
	booksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

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
		case totalBooks, open := <-shelfStatsChan:
			if !open {
				shelfStatsChan = nil
			} else {
				currCrawlStats.TotalBooks = totalBooks
				writeTotalBooksMsg(currCrawlStats, ws)
			}

		case bookFromGoodReads := <-booksFoundFromGoodReadsChan:
			currCrawlStats.BooksCrawled++
			fmt.Printf("[%d](%d) found a book: %+v by %v\n", len(booksFoundFromGoodReadsChan), currCrawlStats.BooksCrawled, bookFromGoodReads.Title, bookFromGoodReads.Author)
			writeGoodReadsBookMsg(bookFromGoodReads, currCrawlStats, ws)
			go thebookshop.SearchForBook(bookFromGoodReads, searchResultsFromTheBookshopChan)

		case searchResultFromTheBookshop := <-searchResultsFromTheBookshopChan:
			currCrawlStats.BooksSearched++
			currCrawlStats.BookMatchFound += len(searchResultFromTheBookshop.TitleMatches)
			// fmt.Printf("%d author and %d title matches for %s\n", len(searchResultFromTheBookshop.AuthorMatches),
			// 	len(searchResultFromTheBookshop.TitleMatches), searchResultFromTheBookshop.SearchBook.Title)
			shouldNotify := db.SaveBookAndNotifyIfNew(searchResultFromTheBookshop)
			if shouldNotify {
				fmt.Printf("*************\n%s: %s is now for sale %s from %s\n\n\n",
					searchResultFromTheBookshop.SearchBook.Author,
					searchResultFromTheBookshop.SearchBook.Title,
					searchResultFromTheBookshop.TitleMatches[0].Price,
					searchResultFromTheBookshop.TitleMatches[0].Link)
			}
			writeSearchResultReturnedMsg(searchResultFromTheBookshop, currCrawlStats, ws)

		}
	}
	fmt.Printf("Exiting. All books queried from Goodreads\n")
	close(booksFoundFromGoodReadsChan)
	close(searchResultsFromTheBookshopChan)
}

func allBooksFound(crawlStats dtos.CrawlStats) bool {
	if ((crawlStats.BooksCrawled == crawlStats.TotalBooks) &&
		(crawlStats.BooksSearched == crawlStats.TotalBooks)) && crawlStats.TotalBooks != -1 {
		return true
	}
	return false
}
