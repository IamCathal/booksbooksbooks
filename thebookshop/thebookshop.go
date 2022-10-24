package thebookshop

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/search"
)

var (
	THE_BOOKSHOP_BASE_URL = "https://thebookshop.ie"
	lastRequestMade       time.Time
	bookshopRequestLock   sync.Mutex
	SLEEP_DURATION        = time.Duration(600 * time.Millisecond)
)

func init() {
	lastRequestMade = time.Now()
}

func SearchForBook(bookInfo dtos.BasicGoodReadsBook, bookSearchResultsChan chan<- dtos.EnchancedSearchResult) dtos.EnchancedSearchResult {
	// startTime := time.Now()
	bookshopRequestLock.Lock()
	for {
		if time.Since(lastRequestMade) > SLEEP_DURATION {
			lastRequestMade = time.Now()
			allBooks := searchTheBookshop(bookInfo, bookSearchResultsChan)
			bookshopRequestLock.Unlock()
			// fmt.Printf("\t\t\t\tWaited %v\n", time.Since(startTime))
			return FindAuthorAndOrTitleMatches(bookInfo, allBooks)
		}
	}
}

func FindAuthorAndOrTitleMatches(bookInfo dtos.BasicGoodReadsBook, searchResult []dtos.TheBookshopBook) dtos.EnchancedSearchResult {
	return search.SearchAllRankFind(bookInfo, searchResult)
}

func searchTheBookshop(bookInfo dtos.BasicGoodReadsBook, bookSearchResultsChan chan<- dtos.EnchancedSearchResult) []dtos.TheBookshopBook {
	searchURL := fmt.Sprintf("%s/search.php?%s", THE_BOOKSHOP_BASE_URL, urlEncodeBookSearch(bookInfo))
	doc, err := goquery.NewDocumentFromReader(getPage(searchURL))
	checkErr(err)
	// fmt.Printf("Search for %s\n", searchURL)
	allBooks := []dtos.TheBookshopBook{}

	doc.Find("ul[class='productGrid']").Each(func(i int, bookReviews *goquery.Selection) {
		bookReviews.Find("li[class='product']").Each(func(k int, bookProduct *goquery.Selection) {

			bookTitle := bookProduct.Find("h4[class='card-title']").Text()
			bookLink, ok := bookProduct.Find("h4[class='card-title'] a").Attr("href")
			if !ok {
				panic(fmt.Sprintf("no bookLink found for %+v with query %s", bookInfo, searchURL))
			}

			bookPrice := bookProduct.Find("span[data-product-price-without-tax='']").Text()
			// fmt.Printf("Title: '%s' Price: '%s' Link: %s\n", bookTitle, bookPrice, bookLink)

			cover, _ := bookProduct.Find("img[class='card-image']").Attr("src")

			author, title := extractAuthorFromTitle(bookTitle)

			foundBook := dtos.TheBookshopBook{
				Title:  title,
				Author: author,
				Price:  bookPrice,
				Link:   bookLink,
				Cover:  cover,
			}
			allBooks = append(allBooks, foundBook)
		})
	})

	// searchResult := make(dtos.AllBookshopBooksSearchResults)
	// returnedBooks := dtos.BookShopBookSearchResult{
	// 	SearchResultBooks: allBooks,
	// }
	// searchResult[bookInfo.Title] = returnedBooks
	// bookSearchResultsChan <- searchResult

	refinedSearchResults := FindAuthorAndOrTitleMatches(bookInfo, allBooks)
	bookSearchResultsChan <- refinedSearchResults

	return allBooks
}

func getPage(pageURL string) io.ReadCloser {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageURL, nil)
	checkErr(err)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Alt-Used", "cdn11.bigcommerce.com")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Host", "cdn11.bigcommerce.com")
	req.Header.Set("TE", "trailers")
	req.Header.Set("Referer", "https://thebookshop.ie/")

	res, err := client.Do(req)
	checkErr(err)
	return res.Body
}
