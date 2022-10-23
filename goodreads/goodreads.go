package goodreads

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	lastRequestMade time.Time
	SLEEP_DURATION  = time.Duration(1 * time.Second)
)

func init() {
	lastRequestMade = time.Now()
}

func GetBooksFromShelf(shelfURL string, shelfStats chan<- int, booksFoundFromGoodReadsChan chan<- dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	if isShelfURL := CheckIsShelfURL(shelfURL); !isShelfURL {
		return []dtos.BasicGoodReadsBook{}
	}
	return extractBooksFromShelfPage(shelfURL, shelfStats, booksFoundFromGoodReadsChan)
}

// func getTotalBooksAndPageSource(shelfURL string) (int, *goquery.Document) {
// 	doc, err := goquery.NewDocumentFromReader(getPage(shelfURL))
// 	checkErr(err)
// 	totalBooks := 0
// 	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
// 		_, totalBooks = extractLoadedCount(loadedCount.Text())
// 	})
// 	return totalBooks, doc
// }

func extractBooksFromShelfPage(shelfURL string, shelfStats chan<- int, booksFoundFromGoodReadsChan chan<- dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	doc, err := goquery.NewDocumentFromReader(getPage(shelfURL))
	checkErr(err)

	allBooks := []dtos.BasicGoodReadsBook{}
	loadedInView := 0
	totalBooks := 0

	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
		loadedInView, totalBooks = extractLoadedCount(loadedCount.Text())
		fmt.Println(loadedInView, totalBooks)
	})

	shelfStats <- totalBooks
	close(shelfStats)

	extractedBooks := extractBooksFromHTML(doc)
	for _, book := range extractedBooks {
		// fmt.Printf("[%d] %d putting in new book %s from page %d\n", len(booksFoundFromGoodReadsChan), i+1, book.Title, 1)
		booksFoundFromGoodReadsChan <- book
	}
	allBooks = append(allBooks, extractedBooks...)

	fmt.Printf("First page done %d/%d books gathered\n", loadedInView, totalBooks)

	if len(allBooks) < totalBooks {
		totalPagesToCrawl := totalPagesToCrawl(totalBooks)
		fmt.Printf("%d pages will need to be crawled\n", totalPagesToCrawl)
		currPageToView := 2
		for {
			if len(allBooks) == totalBooks {
				break
			}
			newUrl := fmt.Sprintf("%s&page=%d", shelfURL, currPageToView)

			newPageDoc, err := goquery.NewDocumentFromReader(getPage(newUrl))
			checkErr(err)

			extractedBooksFromNewPage := extractBooksFromHTML(newPageDoc)
			// fmt.Printf("\n\nGot %d new books from page %d (%s)\n\n", len(extractedBooksFromNewPage), currPageToView, newUrl)
			for _, book := range extractedBooksFromNewPage {
				// fmt.Printf("[%d] Putting in new book %s from page %d\n", len(booksFoundFromGoodReadsChan), book.Title, currPageToView)
				booksFoundFromGoodReadsChan <- book
			}
			allBooks = append(allBooks, extractedBooksFromNewPage...)
			currPageToView++
			sleepIfLongerThanAllotedTimeSinceLastRequest()
		}
	}

	// fmt.Printf("Captured %d books\n", len(allBooks))
	// for i, book := range allBooks {
	// 	fmt.Printf("[%d] %+v\n", i, book)
	// }

	return allBooks
}

func sleepIfLongerThanAllotedTimeSinceLastRequest() {
	fmt.Printf("[goodreads] Time since was %+v Default is %+v\n", time.Since(lastRequestMade), SLEEP_DURATION)
	if time.Since(lastRequestMade) > SLEEP_DURATION {
		lastRequestMade = time.Now()
		fmt.Printf("[goodreads] Was more, not sleeping\n")
		return
	}
	timeDifference := SLEEP_DURATION - time.Since(lastRequestMade)
	fmt.Printf("[goodreads] Was less, sleeping for %+v\n", timeDifference)
	time.Sleep(timeDifference)
	lastRequestMade = time.Now()
}

func extractBooksFromHTML(doc *goquery.Document) []dtos.BasicGoodReadsBook {
	allBooks := []dtos.BasicGoodReadsBook{}
	doc.Find("tbody#booksBody").Each(func(i int, bookReviews *goquery.Selection) {
		bookReviews.Find("tr").Each(func(k int, bookReviewRow *goquery.Selection) {
			title := bookReviewRow.Find("td[class='field title'] a").Text()
			author := bookReviewRow.Find("td[class='field author'] a").Text()
			cover, _ := bookReviewRow.Find("td[class='field cover'] img").Attr("src")
			isbn13 := bookReviewRow.Find("td[class='field isbn13'] div").Text()
			asin := bookReviewRow.Find("td[class='field asin'] div").Text()
			rating := bookReviewRow.Find("td[class='field avg_rating'] div").Text()

			currBook := processBook(title, author, cover, isbn13, asin, rating)
			allBooks = append(allBooks, currBook)
		})
	})
	return allBooks
}

func getPage(pageURL string) io.ReadCloser {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageURL, nil)
	checkErr(err)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "www.goodreads.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", getFakeReferrerPage(pageURL))

	res, err := client.Do(req)
	checkErr(err)
	return res.Body
}
