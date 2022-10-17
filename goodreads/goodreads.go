package goodreads

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func GetBooksFromShelf(shelfURL string) []dtos.BasicGoodReadsBook {
	// is a shelf URL ?

	extractedBooks := extractBooksFromShelfPage(shelfURL)
	return extractedBooks
}

func extractBooksFromShelfPage(shelfURL string) []dtos.BasicGoodReadsBook {
	doc, err := goquery.NewDocumentFromReader(getPage(shelfURL))
	checkErr(err)

	allBooks := []dtos.BasicGoodReadsBook{}
	loadedInView := 0
	totalBooks := 0

	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
		loadedInView, totalBooks = extractLoadedCount(loadedCount.Text())
		fmt.Println(loadedInView, totalBooks)
	})

	allBooks = append(allBooks, extractBooksFromHTML(doc)...)

	fmt.Printf("First page done %d/%d books gathered\n", loadedInView, totalBooks)

	if len(allBooks) < totalBooks {
		fmt.Println("more viewing will be required")
		currPageToView := 2
		for {
			if len(allBooks) == totalBooks {
				break
			}
			newUrl := fmt.Sprintf("%s&page=%d", shelfURL, currPageToView)
			fmt.Printf("Getting new page %s\n", newUrl)

			newPageDoc, err := goquery.NewDocumentFromReader(getPage(newUrl))
			checkErr(err)

			allBooks = append(allBooks, extractBooksFromHTML(newPageDoc)...)
			currPageToView++
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Printf("Captured %d books\n", len(allBooks))
	for i, book := range allBooks {
		fmt.Printf("[%d] %+v\n", i, book)
	}

	return []dtos.BasicGoodReadsBook{}
}

func extractBooksFromHTML(doc *goquery.Document) []dtos.BasicGoodReadsBook {
	allBooks := []dtos.BasicGoodReadsBook{}
	doc.Find("tbody#booksBody").Each(func(i int, bookReviews *goquery.Selection) {
		bookReviews.Find("tr").Each(func(k int, bookReviewRow *goquery.Selection) {

			title := bookReviewRow.Find("td[class='field title'] a").Text()
			author := bookReviewRow.Find("td[class='field author'] a").Text()

			currBook := processBook(stripOfFormatting(title), stripOfFormatting(author))
			allBooks = append(allBooks, currBook)

			// fmt.Printf("%+v\n", currBook)
		})
	})
	return allBooks
}

func getPage(pageURL string) io.ReadCloser {
	res, err := http.Get(pageURL)
	checkErr(err)
	return res.Body
}
