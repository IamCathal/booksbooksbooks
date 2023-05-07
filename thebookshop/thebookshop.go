package thebookshop

import (
	"fmt"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/search"
	"go.uber.org/zap"
)

var (
	logger                *zap.Logger
	THE_BOOKSHOP_BASE_URL = "https://thebookshop.ie"
	lastRequestMade       time.Time
	bookshopRequestLock   sync.Mutex
	SLEEP_DURATION        = time.Duration(600 * time.Millisecond)
)

func init() {
	lastRequestMade = time.Now()
}

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func SearchForBook(bookInfo dtos.BasicGoodReadsBook, bookSearchResultsChan chan<- dtos.EnchancedSearchResult) dtos.EnchancedSearchResult {
	startTime := time.Now()
	bookshopRequestLock.Lock()
	for {
		if time.Since(lastRequestMade) > SLEEP_DURATION {
			lastRequestMade = time.Now()
			logger.Sugar().Infof("Searching for %s by %s", bookInfo.Title, bookInfo.Author)
			bookSearchResults := searchTheBookshop(bookInfo, bookSearchResultsChan)
			bookshopRequestLock.Unlock()
			logger.Sugar().Debugw(fmt.Sprintf("Waited %v before executing TheBookshop.ie search request", time.Since(startTime)),
				zap.String("dignostics", "theBookshopEngine"))
			return bookSearchResults
		}
	}
}

func FindAuthorAndOrTitleMatches(bookInfo dtos.BasicGoodReadsBook, searchResult []dtos.TheBookshopBook) dtos.EnchancedSearchResult {
	return search.SearchAllRankFind(bookInfo, searchResult)
}

func searchTheBookshop(bookInfo dtos.BasicGoodReadsBook, bookSearchResultsChan chan<- dtos.EnchancedSearchResult) dtos.EnchancedSearchResult {
	searchURL := fmt.Sprintf("%s/search.php?%s", THE_BOOKSHOP_BASE_URL, urlEncodeBookSearch(bookInfo))
	doc := goquery.NewDocumentFromNode(controller.Cnt.GetPage(searchURL))
	allBooks := []dtos.TheBookshopBook{}

	doc.Find("ul[class='productGrid']").Each(func(i int, bookReviews *goquery.Selection) {
		bookReviews.Find("li[class='product']").Each(func(k int, bookProduct *goquery.Selection) {
			bookTitle := bookProduct.Find("h4[class='card-title']").Text()
			bookLink, ok := bookProduct.Find("h4[class='card-title'] a").Attr("href")
			if !ok {
				logger.Sugar().Fatalf("no link found on TheBookshop for GoodReads book: %+v with query %s", bookInfo, searchURL)
			}

			bookPrice := bookProduct.Find("span[data-product-price-without-tax='']").Text()
			cover, _ := bookProduct.Find("img[class='card-image']").Attr("src")
			author, title := search.ExtractAuthorFromTheBookShopTitle(bookTitle)
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
	refinedSearchResults := FindAuthorAndOrTitleMatches(bookInfo, allBooks)
	bookSearchResultsChan <- refinedSearchResults
	return refinedSearchResults
}
