package goodreads

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func GetBooksFromShelf(shelfURL string) []dtos.BasicGoodReadsBook {
	// is a shelf URL ?

	extractedBooks := extractBooksFromShelfPage(shelfURL)

	return extractedBooks
}

func extractBooksFromShelfPage(shelfURL string) []dtos.BasicGoodReadsBook {
	doc, err := goquery.NewDocumentFromResponse(getPage(shelfURL))
	checkErr(err)
	allBooks := []dtos.BasicGoodReadsBook{}

	doc.Find("tbody#booksBody").Each(func(i int, bookReviews *goquery.Selection) {
		fmt.Println("found a bookbody")
		bookReviews.Find("tr").Each(func(k int, bookReviewRow *goquery.Selection) {

			title := bookReviewRow.Find("td[class='field title'] a").Text()
			author := bookReviewRow.Find("td[class='field author'] a").Text()
			// fmt.Printf("[%d] '%v' - '%v'\n", k, title, author)

			currBook := processBook(stripOfFormatting(title), stripOfFormatting(author))
			allBooks = append(allBooks, currBook)

			fmt.Printf("%+v\n", currBook)

		})
	})

	return []dtos.BasicGoodReadsBook{}
}

func getPage(pageURL string) *http.Response {
	res, err := http.Get(pageURL)
	checkErr(err)

	return res
}
