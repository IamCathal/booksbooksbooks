package engine

import (
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
)

func init() {
	toBookshopSearchChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchQueryResultsChan := make(chan dtos.BookShopBookSearchResult, 400)
}

func Engine() {
	goodReadsShelfStatsChan := make(chan int, 1)

	go goodreads.GetBooksFromShelf("https://www.goodreads.com/review/list/1753152-sharon?shelf=fantasy", goodReadsShelfStatsChan)

}
