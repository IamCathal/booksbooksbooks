package goodreads

import (
	"fmt"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
)

var (
	BOOKS_DISPLAYED_PER_PAGE = 30
	toBookshopSearchChan     chan<- dtos.BasicGoodReadsBook
)

func HookUpChannels(toBookshopChan chan<- dtos.BasicGoodReadsBook) {
	toBookshopSearchChan = toBookshopChan
}

func Worker(shelfURL string) {
	shelfStatsChan := make(chan int, 1)
	booksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchResultsFromTheBookshopChan := make(chan dtos.AllBookshopBooksSearchResults, 200)

	fmt.Printf("Get the books from shelf\n")
	GetBooksFromShelf(shelfURL, shelfStatsChan, booksFoundFromGoodReadsChan)

	booksFound := 0
	searchResultReturned := 0
	totalBooksFromGoodReads := -1

	for {
		if allBooksFound(booksFound, searchResultReturned, totalBooksFromGoodReads) {
			break
		}

		select {
		case totalBooks := <-shelfStatsChan:
			fmt.Printf("[]][[]][][][][][][ %d total books\n\n", totalBooks)
			totalBooksFromGoodReads = totalBooks
		case bookFromGoodReads := <-booksFoundFromGoodReadsChan:
			fmt.Printf("(%d) found a book: %+v\n", booksFound, bookFromGoodReads)
			booksFound++
			thebookshop.SearchForBook(bookFromGoodReads, searchResultsFromTheBookshopChan)
		case searchResultFromTheBookshop := <-searchResultsFromTheBookshopChan:
			fmt.Printf("\nSearch result found: %+v\n\n", searchResultFromTheBookshop)
			searchResultReturned++
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
