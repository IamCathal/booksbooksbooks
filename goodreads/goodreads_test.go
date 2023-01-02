package goodreads

import (
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

var (
	validShelfURL             = "https://www.goodreads.com/review/list/26367680?shelf=read"
	stephenKingGoodreadsShelf = ""
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)
	db.SetLogger(logger)
	db.ConnectToRedis()
	db.SetTestDataIdentifiers()
	stephenKingGoodreadsShelf = readFile("../testData/stephenKingGoodreadsShelfOneBook.html")

	code := m.Run()

	os.Exit(code)
}

func TestCheckIsShelfURL(t *testing.T) {
	assert.True(t, CheckIsShelfURL(validShelfURL))
}

func TestCheckIsShelfURLCatchesAnInvalidShelfURL(t *testing.T) {
	invalidShelfURL := "https://www.goodreads.com/review/list/26367680"
	assert.False(t, CheckIsShelfURL(invalidShelfURL))
}

func TestStripOfFormatting(t *testing.T) {
	assert.Equal(t, "charlie byrnes", stripOfFormatting("\n\ncharlie byrne\n\n\ns\n\n"))
}

func TestExtractTitleDetailsIfPossibleWithNoSeries(t *testing.T) {
	expectedTitle := "Billy Summers"
	expectedSeriesText := ""

	actualTitle, actualSeriesTitle := extractTitleDetailsIfPossible("Billy Summers")

	assert.Equal(t, expectedTitle, actualTitle)
	assert.Equal(t, expectedSeriesText, actualSeriesTitle)
}

func TestExtractTitleDetailsIfPossibleWithSeries(t *testing.T) {
	expectedTitle := "The Name of the Wind"
	expectedSeriesText := "(The Kingkiller Chronicle, #1)"

	actualTitle, actualSeriesTitle := extractTitleDetailsIfPossible("The Name of the Wind           (The Kingkiller Chronicle, #1)")

	assert.Equal(t, expectedTitle, actualTitle)
	assert.Equal(t, expectedSeriesText, actualSeriesTitle)
}

func TestExtractLoadedCount(t *testing.T) {
	loadedTextRaw := `
	30 of 89 loaded
  	`
	expectedLoaded := 30
	expectedTotal := 89

	actualLoaded, actualTotal := extractLoadedCount(loadedTextRaw)

	assert.Equal(t, expectedLoaded, actualLoaded)
	assert.Equal(t, expectedTotal, actualTotal)
}

func TestTotalPagesToCrawlForLessThanSeperator(t *testing.T) {
	assert.Equal(t, 1, totalPagesToCrawl(BOOK_COUNT_PER_PAGE-1))
}

func TestTotalPagesToCrawlForEqualToSeperator(t *testing.T) {
	assert.Equal(t, 1, totalPagesToCrawl(BOOK_COUNT_PER_PAGE))
}

func TestTotalPagesToCrawlOneMoreThanSeperator(t *testing.T) {
	assert.Equal(t, 2, totalPagesToCrawl(BOOK_COUNT_PER_PAGE+1))
}

func TestTotalPagesToCrawlFiveTimesMoreThanSeperator(t *testing.T) {
	assert.Equal(t, 5, totalPagesToCrawl(BOOK_COUNT_PER_PAGE*5))
}

func TestExtractPureTitle(t *testing.T) {
	expectedTitle := "The Name Of The Wind"

	assert.Equal(t, expectedTitle, extractPureTitle("The Name Of The Wind (King Killer Chronicle #1)"))
}

func TestExtractPureTitleWithNoSeriesReturnsTheSameThing(t *testing.T) {
	expectedTitle := "The Name Of The Wind"

	assert.Equal(t, expectedTitle, extractPureTitle("The Name Of The Wind"))
}

func TestEnsureAllAverageRatingsAreOfTypeString(t *testing.T) {
	jsonString := `[{"imageUrl":"https://i.gr-assets.com/images/S/compressed.photo.goodreads.com/books/1183241465i/1393636._SY75_.jpg","bookId":"1393636","workId":"3634570","bookUrl":"/book/show/1393636.Le_Messie_de_Dune","from_search":true,"from_srp":true,"qid":"EZCLG1pYwp","rank":1,"title":"Le Messie de Dune","bookTitleBare":"Le Messie de Dune","numPages":316,"avgRating":0,"ratingsCount":220195,"author":{"id":58,"name":"Frank Herbert","isGoodreadsAuthor":false,"profileUrl":"https://www.goodreads.com/author/show/58.Frank_Herbert","worksListUrl":"https://www.goodreads.com/author/list/58.Frank_Herbert"},"kcrPreviewUrl":null,"description":{"html":"<b>This is an alternate cover edition for ISBN 978-2-266-15451-2.</b><br/><br/>Paul Atréides a triomphé de ses ennemis. En douze ans de guerre sainte, ses Fremen ont conquis l&apos;univers. Il est …","truncated":true,"fullContentUrl":"https://www.goodreads.com/book/show/1393636.Le_Messie_de_Dune"}}]`

	booksFoundRes := []dtos.GoodReadsSearchBookResult{}
	jsonWithFixedAvgRating := ensureAllAverageRatingsAreOfTypeString([]byte(jsonString))

	err := json.Unmarshal(jsonWithFixedAvgRating, &booksFoundRes)
	if err != nil {
		assert.Fail(t, "expect to now throw an error unmarshaling fixed json")
	}

	assert.Equal(t, "0.0", booksFoundRes[0].AvgRating)
}

func TestSleepIfLongerThanAllotedTimeSinceLastRequest(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)

	SLEEP_DURATION = time.Duration(5 * time.Millisecond)
	mockController.On("Sleep", mock.AnythingOfType("time.Duration")).After(SLEEP_DURATION).Return()
	startTime := time.Now()

	sleepIfLongerThanAllotedTimeSinceLastRequest()

	timeTaken := time.Since(startTime)
	assert.GreaterOrEqual(t, timeTaken.Abs().Milliseconds(), SLEEP_DURATION.Abs().Milliseconds())
}

func TestSleepIfLongerThanAllotedTimeSinceLastRequestDoesnSleepWhenTimeIsLessThanDelay(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	lastRequestMade = time.Now().Add(-time.Duration(time.Second * 20))

	startTime := time.Now()

	sleepIfLongerThanAllotedTimeSinceLastRequest()

	timeTaken := time.Since(startTime)
	assert.Less(t, timeTaken.Abs().Milliseconds(), SLEEP_DURATION.Abs().Milliseconds())
}

func TestProcessBookStripsOfAllFormatting(t *testing.T) {
	expectedID := ksuid.New().String()
	expectedProcessedBook := dtos.BasicGoodReadsBook{
		ID:         expectedID,
		Title:      "The Name Of The Wind",
		Author:     "Patrick Rothfuss",
		SeriesText: "The Kingkiller Chronicle #1",
		Link:       GOODREADS_BASE_BOOK_URL + "bookName",
		Cover:      "coverLink",
		Isbn13:     "isbn13",
		Asin:       "asin",
		Rating:     3.2,
	}

	actualProcessedBook := processBook("          The Name Of The Wind           The Kingkiller Chronicle\n\n\n #1\n",
		"\n\n\n\n\nPatrick Rothfuss",
		"    cov\nerLink",
		"isbn1\n\n\n\n3",
		"\n\n                          asin",
		"3.2",
		"bookName")

	assert.Equal(t, expectedProcessedBook.Title, actualProcessedBook.Title)
	assert.Equal(t, expectedProcessedBook.Author, actualProcessedBook.Author)
	assert.Equal(t, expectedProcessedBook.SeriesText, actualProcessedBook.SeriesText)
	assert.Equal(t, expectedProcessedBook.Link, actualProcessedBook.Link)
	assert.Equal(t, expectedProcessedBook.Cover, actualProcessedBook.Cover)
	assert.Equal(t, expectedProcessedBook.Isbn13, actualProcessedBook.Isbn13)
	assert.Equal(t, expectedProcessedBook.Asin, actualProcessedBook.Asin)
	assert.Equal(t, expectedProcessedBook.Rating, actualProcessedBook.Rating)
}

func TestExtractBooksFromHTML(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	resetDBFields()

	expectedBooks := []dtos.BasicGoodReadsBook{
		{
			Title:  "Doing Harm",
			Author: "Parsons, Kelly",
			Link:   "https://www.goodreads.com/book/show/17934474-doing-harm",
			Cover:  "https://i.gr-assets.com/images/S/compressed.photo.goodreads.com/books/1400983374l/17934474._SY75_.jpg",
		},
	}

	mockController.On("GetPage", mock.AnythingOfType("string")).Return(getHtmlNode(stephenKingGoodreadsShelf))

	doc := goquery.NewDocumentFromNode(controller.Cnt.GetPage(""))
	actualExtractedBooks := extractBooksFromHTML(doc)

	assert.Equal(t, expectedBooks[0].Title, actualExtractedBooks[0].Title)
	assert.Equal(t, expectedBooks[0].Author, actualExtractedBooks[0].Author)
	assert.Equal(t, expectedBooks[0].Link, actualExtractedBooks[0].Link)
	assert.Equal(t, expectedBooks[0].Cover, actualExtractedBooks[0].Cover)
}

func TestGetAvailableBooksFromSearchResultWithAddMoreAuthorBooksDisabledOnlyAddsTitleMatches(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	resetDBFields()
	db.SetAddMoreAuthorBooksToAvailableBooksList(false)

	searchBook := dtos.BasicGoodReadsBook{
		Author: "Stephen Lawhead",
		Title:  "Empryion",
	}
	searchResults := []dtos.EnchancedSearchResult{
		{
			SearchBook: searchBook,
			TitleMatches: []dtos.TheBookshopBook{
				{
					Author: searchBook.Author,
					Title:  searchBook.Title,
				},
			},
			AuthorMatches: []dtos.TheBookshopBook{
				{
					Author: searchBook.Author,
					Title:  "Another Lawhead Book",
				},
			},
		},
	}

	availableBooksFromSearchResult := GetAvailableBooksFromSearchResult(searchResults)

	assert.Equal(t, 1, len(availableBooksFromSearchResult))
}

func TestGetAvailableBooksFromSearchResultWithAddMoreAuthorBooksEnabledAddsTitleAndAuthorMatches(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	resetDBFields()
	db.SetAddMoreAuthorBooksToAvailableBooksList(true)

	searchBook := dtos.BasicGoodReadsBook{
		Author: "Stephen Lawhead",
		Title:  "Empryion",
	}
	searchResults := []dtos.EnchancedSearchResult{
		{
			SearchBook: searchBook,
			TitleMatches: []dtos.TheBookshopBook{
				{
					Author: searchBook.Author,
					Title:  searchBook.Title,
				},
			},
			AuthorMatches: []dtos.TheBookshopBook{
				{
					Author: searchBook.Author,
					Title:  "Another Lawhead Book",
				},
			},
		},
	}

	availableBooksFromSearchResult := GetAvailableBooksFromSearchResult(searchResults)

	assert.Equal(t, 2, len(availableBooksFromSearchResult))
}

func TestFilterSeriesTitleFromSeriesText(t *testing.T) {
	assert.Equal(t, "Dune", FilterSeriesTitleFromSeriesText("(Dune #3)"))
}

func TestFilterSeriesTitleFromSeriesTextWithDotInTitleSequence(t *testing.T) {
	assert.Equal(t, "Sorcery Ascendant", FilterSeriesTitleFromSeriesText("(Sorcery Ascendant, #0.5)"))
}

func TestExtractCommunityRatingElements(t *testing.T) {
	expectedRating := float32(3.83)
	expectedPublishedYear := 2022

	actualRating, actualPublishedYear := extractCommunityRatingElementsFromText("3.83 · 332239 Ratings · 40480 Reviews · published 2022 · 151 editions")

	assert.Equal(t, expectedPublishedYear, actualPublishedYear)
	assert.Equal(t, expectedRating, actualRating)
}

func resetDBFields() {
	db.SetKnownAuthors([]dtos.KnownAuthor{})
	db.SetAddMoreAuthorBooksToAvailableBooksList(false)
	db.SetAvailableBooks([]dtos.AvailableBook{})
}

func getHtmlNode(webpageStr string) *html.Node {
	htmlNodeResponse, err := html.Parse(strings.NewReader(webpageStr))
	if err != nil {
		panic(err)
	}
	return htmlNodeResponse
}

func readFile(fileName string) string {
	data, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}
	return string(data)
}
