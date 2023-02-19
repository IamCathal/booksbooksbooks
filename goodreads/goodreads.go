package goodreads

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

var (
	logger             *zap.Logger
	lastRequestMade    time.Time
	SLEEP_DURATION     = time.Duration(1 * time.Second)
	GOODREADS_BASE_URL = "https://goodreads.com"
)

func init() {
	lastRequestMade = time.Now()
}

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func GetBooksFromShelves(shelveURLs []string, shelfStats chan<- int, booksFoundFromGoodReadsChan chan<- dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	booksFromAllShelves := []dtos.BasicGoodReadsBook{}
	for _, shelfToCrawl := range shelveURLs {
		fmt.Println("get books from " + shelfToCrawl)
		booksFromAllShelves = append(booksFromAllShelves, GetBooksFromShelf(shelfToCrawl, shelfStats, booksFoundFromGoodReadsChan)...)
	}
	fmt.Println("done getting all books from shelves")
	close(shelfStats)
	return booksFromAllShelves
}

func GetBooksFromShelf(shelfURL string, shelfStats chan<- int, booksFoundFromGoodReadsChan chan<- dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	if isShelfURL := CheckIsShelfURL(shelfURL); !isShelfURL {
		return []dtos.BasicGoodReadsBook{}
	}
	return extractBooksFromShelfPage(shelfURL, shelfStats, booksFoundFromGoodReadsChan)
}

func extractBooksFromShelfPage(shelfURL string, shelfStats chan<- int, booksFoundFromGoodReadsChan chan<- dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	doc := goquery.NewDocumentFromNode(controller.Cnt.GetPage(shelfURL))

	allBooks := []dtos.BasicGoodReadsBook{}
	totalBooks := 0

	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
		_, totalBooks = extractLoadedCount(loadedCount.Text())
		logger.Sugar().Infof("Shelf %s has %d total books to crawl", shelfURL, totalBooks)
	})

	shelfStats <- totalBooks
	extractedBooks := extractBooksFromHTML(doc)
	for _, book := range extractedBooks {
		booksFoundFromGoodReadsChan <- book
	}
	allBooks = append(allBooks, extractedBooks...)

	logger.Sugar().Infof("Extracted all %d books on page 1", len(extractedBooks))

	if len(allBooks) < totalBooks {
		totalPagesToCrawl := totalPagesToCrawl(totalBooks)
		logger.Sugar().Infof("Shelf had >%d books, %d pages will need to be crawled", BOOK_COUNT_PER_PAGE, totalPagesToCrawl)
		currPageToView := 2
		for {
			if len(allBooks) == totalBooks {
				break
			}
			newUrl := fmt.Sprintf("%s&page=%d", shelfURL, currPageToView)
			newPageDoc := goquery.NewDocumentFromNode(controller.Cnt.GetPage(newUrl))

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

func GenerateShelfToCrawlEntryAndSave(shelfURL string) dtos.ShelfToCrawl {
	shelfToSave := GenerateShelfToCrawlEntry(shelfURL)
	db.AddShelfToShelvesToCrawl(shelfToSave)
	return shelfToSave
}

func GenerateShelfToCrawlEntry(shelfURL string) dtos.ShelfToCrawl {
	doc := goquery.NewDocumentFromNode(controller.Cnt.GetPage(shelfURL))
	totalBooks := 0

	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
		_, totalBooks = extractLoadedCount(loadedCount.Text())
	})
	extractedBooks := extractBooksFromHTML(doc)

	return dtos.ShelfToCrawl{
		CrawlKey:  db.GetKeyForRecentCrawlBreadcrumb(shelfURL),
		ShelfURL:  shelfURL,
		BookCount: totalBooks,
		Covers:    getFirstNBookCovers(extractedBooks, 30),
	}
}

func GetPreviewForShelf(shelfURL string) ([]dtos.BasicGoodReadsBook, int) {
	doc := goquery.NewDocumentFromNode(controller.Cnt.GetPage(shelfURL))
	totalBooks := 0

	doc.Find("div[id='infiniteStatus']").Each(func(i int, loadedCount *goquery.Selection) {
		_, totalBooks = extractLoadedCount(loadedCount.Text())
		logger.Sugar().Infof("Shelf %s has %d total books to crawl", shelfURL, totalBooks)
	})

	extractedBooks := extractBooksFromHTML(doc)

	if len(extractedBooks) >= 12 {
		return extractedBooks[:12], totalBooks
	}
	return extractedBooks, totalBooks
}

func SearchGoodreads(bookPurchaseInfo dtos.TheBookshopBook) (bool, dtos.BasicGoodReadsBook) {
	bookSearchName := fmt.Sprintf("%s %s", bookPurchaseInfo.Author, extractPureTitle(bookPurchaseInfo.Title))
	body := controller.Cnt.Get(fmt.Sprintf("https://www.goodreads.com/book/auto_complete?format=json&q=%s", url.QueryEscape(bookSearchName)))

	booksFoundRes := []dtos.GoodReadsSearchBookResult{}
	body = ensureAllAverageRatingsAreOfTypeString(body)

	err := json.Unmarshal(body, &booksFoundRes)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	if len(booksFoundRes) == 0 {
		logger.Sugar().Infof("Could not find goodreads book for: %+v", bookPurchaseInfo)
		return false, dtos.BasicGoodReadsBook{}
	}

	topSearchResult := dtos.BasicGoodReadsBook{
		ID:         ksuid.New().String(),
		Title:      booksFoundRes[0].BookTitleBare,
		Author:     booksFoundRes[0].Author.Name,
		SeriesText: booksFoundRes[0].Title[len(booksFoundRes[0].BookTitleBare):],
		Link:       booksFoundRes[0].Description.FullContentURL,
		Cover:      booksFoundRes[0].ImageURL,
		// Isbn13:  ,
		// Asin:    ,
		Rating: float32(strToFloat(booksFoundRes[0].AvgRating)),
	}
	return true, topSearchResult
}

func GetSeriesLink(bookInSeries dtos.BasicGoodReadsBook) string {
	individualBookPage := controller.Cnt.GetPage(bookInSeries.Link)
	return getSeriesLink(individualBookPage)
}

func GetSeriesDetailsFromLink(seriesLink string, seriesDetailsChan chan<- dtos.Series) dtos.Series {
	seriesInfo := extractSeriesInfo(seriesLink)
	seriesDetailsChan <- seriesInfo
	return seriesInfo
}

func GetSeriesDetails(bookInSeries dtos.BasicGoodReadsBook) dtos.Series {
	return extractSeriesInfo(GetSeriesLink(bookInSeries))
}
