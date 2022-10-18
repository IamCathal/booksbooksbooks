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
	SLEEP_DURATION = time.Duration(1 * time.Second)
)

func GetBooksFromShelf(shelfURL string) []dtos.BasicGoodReadsBook {
	if isShelfURL := checkIsShelfURL(shelfURL); !isShelfURL {
		return []dtos.BasicGoodReadsBook{}
	}
	return extractBooksFromShelfPage(shelfURL)
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
		totalPagesToCrawl := totalPagesToCrawl(totalBooks)
		fmt.Printf("%d pages will need to be crawled\n", totalPagesToCrawl)
		currPageToView := 2
		for {
			if len(allBooks) == totalBooks {
				break
			}
			newUrl := fmt.Sprintf("%s&page=%d", shelfURL, currPageToView)
			fmt.Printf("[%d/%d] Getting new page %s\n", currPageToView, totalPagesToCrawl, newUrl)

			newPageDoc, err := goquery.NewDocumentFromReader(getPage(newUrl))
			checkErr(err)

			allBooks = append(allBooks, extractBooksFromHTML(newPageDoc)...)
			currPageToView++
			time.Sleep(SLEEP_DURATION)
		}
	}

	fmt.Printf("Captured %d books\n", len(allBooks))
	for i, book := range allBooks {
		fmt.Printf("[%d] %+v\n", i, book)
	}

	return allBooks
}

func extractBooksFromHTML(doc *goquery.Document) []dtos.BasicGoodReadsBook {
	allBooks := []dtos.BasicGoodReadsBook{}
	doc.Find("tbody#booksBody").Each(func(i int, bookReviews *goquery.Selection) {
		bookReviews.Find("tr").Each(func(k int, bookReviewRow *goquery.Selection) {
			title := bookReviewRow.Find("td[class='field title'] a").Text()
			author := bookReviewRow.Find("td[class='field author'] a").Text()

			currBook := processBook(stripOfFormatting(title), stripOfFormatting(author))
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
