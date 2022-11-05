package goodreads

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"go.uber.org/zap"
)

var (
	logger          *zap.Logger
	lastRequestMade time.Time
	SLEEP_DURATION  = time.Duration(1 * time.Second)
)

func init() {
	lastRequestMade = time.Now()
}

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func GetBooksFromShelf(shelfURL string, shelfStats chan<- int, booksFoundFromGoodReadsChan chan<- dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	if isShelfURL := CheckIsShelfURL(shelfURL); !isShelfURL {
		return []dtos.BasicGoodReadsBook{}
	}
	return extractBooksFromShelfPage(shelfURL, shelfStats, booksFoundFromGoodReadsChan)
}

func extractBooksFromShelfPage(shelfURL string, shelfStats chan<- int, booksFoundFromGoodReadsChan chan<- dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	doc, err := goquery.NewDocumentFromReader(getPage(shelfURL))
	checkErr(err)

	allBooks := []dtos.BasicGoodReadsBook{}
	totalBooks := 0

	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
		_, totalBooks = extractLoadedCount(loadedCount.Text())
		logger.Sugar().Infof("Shelf %s has %d total books to crawl", shelfURL, totalBooks)
	})

	shelfStats <- totalBooks
	close(shelfStats)
	extractedBooks := extractBooksFromHTML(doc)
	for _, book := range extractedBooks {
		booksFoundFromGoodReadsChan <- book
	}
	allBooks = append(allBooks, extractedBooks...)

	logger.Sugar().Infof("Extracted all %d books on page 1\n", len(extractedBooks))

	if len(allBooks) < totalBooks {
		totalPagesToCrawl := totalPagesToCrawl(totalBooks)
		logger.Sugar().Info("Shelf had >%d books, %d pages will need to be crawled", BOOK_COUNT_PER_PAGE, totalPagesToCrawl)
		currPageToView := 2
		for {
			if len(allBooks) == totalBooks {
				break
			}
			newUrl := fmt.Sprintf("%s&page=%d", shelfURL, currPageToView)

			newPageDoc, err := goquery.NewDocumentFromReader(getPage(newUrl))
			checkErr(err)

			extractedBooksFromNewPage := extractBooksFromHTML(newPageDoc)
			for _, book := range extractedBooksFromNewPage {
				booksFoundFromGoodReadsChan <- book
			}
			allBooks = append(allBooks, extractedBooksFromNewPage...)
			currPageToView++
			sleepIfLongerThanAllotedTimeSinceLastRequest()
		}
	}
	return allBooks
}

func sleepIfLongerThanAllotedTimeSinceLastRequest() {
	logger.Sugar().Debugw(fmt.Sprintf("Time since last goodreads request was %+v Default is %+v", time.Since(lastRequestMade), SLEEP_DURATION),
		zap.String("dignostics", "goodReadsEngine"))
	if time.Since(lastRequestMade) > SLEEP_DURATION {
		lastRequestMade = time.Now()
		logger.Sugar().Debugw(fmt.Sprintf("[goodreads] Time since last request more than %d, not sleeping", SLEEP_DURATION),
			zap.String("dignostics", "goodReadsEngine"))
		return
	}
	timeDifference := SLEEP_DURATION - time.Since(lastRequestMade)
	logger.Sugar().Debugw(fmt.Sprintf("Time since last request was less than %d, sleeping for %+v", SLEEP_DURATION, timeDifference),
		zap.String("dignostics", "goodReadsEngine"))
	time.Sleep(timeDifference)
	lastRequestMade = time.Now()
}

func GetBookCountForShelf(shelfURL string) int {
	doc, err := goquery.NewDocumentFromReader(getPage(shelfURL))
	checkErr(err)
	totalBooks := 0

	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
		_, totalBooks = extractLoadedCount(loadedCount.Text())
	})

	return totalBooks
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
			link, _ := bookReviewRow.Find("td[class='field title'] a").Attr("href")

			currBook := processBook(title, author, cover, isbn13, asin, rating, link)
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
