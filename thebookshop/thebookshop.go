package thebookshop

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	THE_BOOKSHOP_BASE_URL = "https://thebookshop.ie"
	lastRequestMade       time.Time
	SLEEP_DURATION        = time.Duration(1 * time.Second)
)

func init() {
	lastRequestMade = time.Now()
}

func SearchForBooks(searchBooks []dtos.BasicGoodReadsBook) dtos.AllBookshopBooksSearchResults {
	searchResults := make(dtos.AllBookshopBooksSearchResults)

	// for _, bookToSearch := range searchBooks {
	// 	go searchForBookWithThrottling(bookToSearch)
	// }

	// for _, bookInfo := range searchBooks {
	// 	searchResults[bookInfo.Title] = SearchForBook(bookInfo)
	// }

	return searchResults
}

func SearchForBook(bookInfo dtos.BasicGoodReadsBook, bookSearchResultsChan chan<- dtos.AllBookshopBooksSearchResults) dtos.BookShopBookSearchResult {
	allBooks := searchTheBookshop(bookInfo, bookSearchResultsChan)
	// for i, book := range allBooks {
	// 	fmt.Printf("[%d] %+v\n", i, book)
	// }
	sleepIfLongerThanAllotedTimeSinceLastRequest()
	return dtos.BookShopBookSearchResult{
		SearchResultBooks: allBooks,
	}
}

func searchTheBookshop(bookInfo dtos.BasicGoodReadsBook, bookSearchResultsChan chan<- dtos.AllBookshopBooksSearchResults) []dtos.TheBookshopBook {
	searchURL := fmt.Sprintf("%s/search.php?%s", THE_BOOKSHOP_BASE_URL, urlEncodeBookSearch(bookInfo))
	doc, err := goquery.NewDocumentFromReader(getPage(searchURL))
	checkErr(err)
	fmt.Printf("Search for %s\n", searchURL)
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

	searchResult := make(dtos.AllBookshopBooksSearchResults)
	returnedBooks := dtos.BookShopBookSearchResult{
		SearchResultBooks: allBooks,
	}
	searchResult[bookInfo.Title] = returnedBooks
	bookSearchResultsChan <- searchResult

	return allBooks
}

func sleepIfLongerThanAllotedTimeSinceLastRequest() {
	fmt.Printf("[thebookshop] Time since was %+v Default is %+v\n", time.Since(lastRequestMade), SLEEP_DURATION)
	if time.Since(lastRequestMade) > SLEEP_DURATION {
		lastRequestMade = time.Now()
		fmt.Printf("[thebookshop] Was more, not sleeping\n")
		return
	}
	timeDifference := SLEEP_DURATION - time.Since(lastRequestMade)
	fmt.Printf("[thebookshop] Was less, sleeping for %+v\n", timeDifference)
	time.Sleep(timeDifference)
	lastRequestMade = time.Now()
}

func getPage(pageURL string) io.ReadCloser {
	fmt.Println(pageURL)
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
