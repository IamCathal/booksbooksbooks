package engine

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
)

var (
	BOOKS_DISPLAYED_PER_PAGE = 30
	toBookshopSearchChan     chan<- dtos.BasicGoodReadsBook
)

func HookUpChannels(toBookshopChan chan<- dtos.BasicGoodReadsBook) {
	toBookshopSearchChan = toBookshopChan
}

func Worker(shelfURL string, ws *websocket.Conn) {
	shelfStatsChan := make(chan int, 1)
	booksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchResultsFromTheBookshopChan := make(chan dtos.AllBookshopBooksSearchResults, 200)

	fmt.Printf("Get the books from shelf\n")
	goodreads.GetBooksFromShelf(shelfURL, shelfStatsChan, booksFoundFromGoodReadsChan)

	booksFound := 0
	searchResultReturned := 0
	totalBooksFromGoodReads := -1

	for {
		if allBooksFound(booksFound, searchResultReturned, totalBooksFromGoodReads) {
			break
		}

		select {
		case totalBooks := <-shelfStatsChan:
			totalBooksFromGoodReads = totalBooks
			totalBooksMsg := dtos.WsTotalBooks{
				TotalBooks: totalBooks,
				CrawlStats: dtos.CrawlStats{
					TotalBooks:    totalBooks,
					BooksCrawled:  booksFound,
					BooksSearched: searchResultReturned,
				},
			}
			jsonStr, err := json.Marshal(totalBooksMsg)
			if err != nil {
				panic(err)
			}
			WriteMsg(jsonStr, ws)

		case bookFromGoodReads := <-booksFoundFromGoodReadsChan:
			booksFound++
			fmt.Printf("(%d) found a book: %+v\n", booksFound, bookFromGoodReads)
			goodReadsBookMsg := dtos.WsGoodreadsBook{
				BookInfo: bookFromGoodReads,
				CrawlStats: dtos.CrawlStats{
					TotalBooks:    totalBooksFromGoodReads,
					BooksCrawled:  booksFound,
					BooksSearched: searchResultReturned,
				},
			}
			jsonStr, err := json.Marshal(goodReadsBookMsg)
			if err != nil {
				panic(err)
			}
			WriteMsg(jsonStr, ws)
			thebookshop.SearchForBook(bookFromGoodReads, searchResultsFromTheBookshopChan)

		case searchResultFromTheBookshop := <-searchResultsFromTheBookshopChan:
			searchResultReturned++
			fmt.Printf("\nSearch result found: %+v\n\n", searchResultFromTheBookshop)
			searchResultMsg := dtos.WsBookshopSearchResult{
				Result: searchResultFromTheBookshop,
				CrawlStats: dtos.CrawlStats{
					TotalBooks:    totalBooksFromGoodReads,
					BooksCrawled:  booksFound,
					BooksSearched: searchResultReturned,
				},
			}
			jsonStr, err := json.Marshal(searchResultMsg)
			if err != nil {
				panic(err)
			}
			WriteMsg(jsonStr, ws)

		}
	}
	fmt.Printf("Exiting. All books queried from Goodreads")
}

func allBooksFound(booksFound, searchResultsReturned, total int) bool {
	if ((booksFound == total) && (searchResultsReturned == total)) && total != -1 {
		return true
	}
	return false
}

func WriteMsg(msg []byte, ws *websocket.Conn) {
	err := ws.WriteMessage(1, msg)
	if err != nil {
		panic(err)
	}
}
