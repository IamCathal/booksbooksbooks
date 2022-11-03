package engine

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
	"github.com/iamcathal/booksbooksbooks/util"
	"go.uber.org/zap"
)

var (
	logger                   *zap.Logger
	BOOKS_DISPLAYED_PER_PAGE = 30
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func Worker(shelfURL string, ws *websocket.Conn) {
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfURL '%s' given", shelfURL)
		writeErrorMsg(errorMsg, ws)
		return
	}

	db.SaveRecentCrawlStats(shelfURL)
	previouslyKnownAvailableBooks := db.GetAvailableBooksMap()

	shelfStatsChan := make(chan int, 1)
	booksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

	logger.Sugar().Infof("Retrieving books from shelf: %s\n", shelfURL)
	goodreads.GetBooksFromShelf(shelfURL, shelfStatsChan, booksFoundFromGoodReadsChan)

	booksFound := 0
	searchResultsReturned := 0
	totalBooksFromGoodReads := -1
	currCrawlStats := dtos.CrawlStats{
		TotalBooks:    totalBooksFromGoodReads,
		BooksCrawled:  booksFound,
		BooksSearched: searchResultsReturned,
	}
	newBooksFound := 0
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
			logger.Sugar().Infof("[booksFound: %d][booksCrawled: %d] Found a GoodReads book: %+v by %v",
				len(booksFoundFromGoodReadsChan), currCrawlStats.BooksCrawled,
				bookFromGoodReads.Title, bookFromGoodReads.Author)
			writeGoodReadsBookMsg(bookFromGoodReads, currCrawlStats, ws)
			go thebookshop.SearchForBook(bookFromGoodReads, searchResultsFromTheBookshopChan)

		case searchResultFromTheBookshop := <-searchResultsFromTheBookshopChan:
			currCrawlStats.BooksSearched++
			currCrawlStats.BookMatchFound += len(searchResultFromTheBookshop.TitleMatches)
			if len(searchResultFromTheBookshop.TitleMatches) > 0 {
				// TOOD handle multiple title searches
				if bookIsNew := bookIsNew(searchResultFromTheBookshop.TitleMatches[0], previouslyKnownAvailableBooks); bookIsNew {
					newBooksFound++
					logger.Sugar().Infof("Found a book that's for sale: %s by %s for %s at %s",
						searchResultFromTheBookshop.SearchBook.Title,
						searchResultFromTheBookshop.SearchBook.Author,
						searchResultFromTheBookshop.TitleMatches[0].Price,
						searchResultFromTheBookshop.TitleMatches[0].Link)
					writeNewAvailableBookMsg(searchResultFromTheBookshop.TitleMatches[0], currCrawlStats, ws)
					newBook := dtos.AvailableBook{
						BookInfo:         searchResultFromTheBookshop.SearchBook,
						BookPurchaseInfo: searchResultFromTheBookshop.TitleMatches[0],
					}
					db.AddAvailableBook(newBook)
					for _, book := range searchResultFromTheBookshop.TitleMatches {
						util.SendNewBookIsAvailableMessage(book)
					}
				}
			}
			writeSearchResultReturnedMsg(searchResultFromTheBookshop, currCrawlStats, ws)
		}
	}

	logger.Sugar().Infof("Finished. Crawled %d books from GoodReads and made %d searches to TheBookshop.ie which had %d new books",
		currCrawlStats.BooksCrawled, currCrawlStats.BooksSearched, newBooksFound)
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
