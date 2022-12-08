package engine

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/net/html"
	"gotest.tools/assert"
)

var (
	validShelfURL = "https://www.goodreads.com/review/list/26367680?shelf=read"

	// eee
	stephenKingGoodreadsShelf              = ""
	stephenKingGoodreadsShelfOneBook       = ""
	parsonsKellyDoingHarmTheBookshopSearch = ""
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
	goodreads.SetLogger(logger)
	thebookshop.SetLogger(logger)
	thebookshop.SLEEP_DURATION = time.Duration(100 * time.Nanosecond)
	goodreads.SLEEP_DURATION = time.Duration(100 * time.Nanosecond)

	db.ConnectToRedis()
	db.SetTestDataIdentifiers()
	loadMockSearchResults()

	code := m.Run()

	os.Exit(code)
}

func loadMockSearchResults() {
	stephenKingGoodreadsShelf = readFile("testData/stephenKingGoodreadsShelf.html")
	stephenKingGoodreadsShelfOneBook = readFile("testData/stephenKingGoodreadsShelfOneBook.html")
	parsonsKellyDoingHarmTheBookshopSearch = readFile("testData/parsonsKellyDoingHarmTheBookshop.html")
}

func TestWorker(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	resetDBFields()

	mockController.On("Sleep", mock.Anything).After(100 * time.Nanosecond).Return()

	// Get the goodreads page
	mockController.On("GetPage", validShelfURL).Once().Return(getHtmlNode(stephenKingGoodreadsShelfOneBook))
	// Don't bother with validating websocket messages just yet
	mockController.On("WriteWsMessage", mock.Anything, mock.AnythingOfType(("*websocket.Conn")), mock.Anything).Return(nil)
	mockController.On("DeliverWebhook", mock.AnythingOfType("dtos.DiscordMsg")).Return(nil)

	mockController.On("GetPage", "https://thebookshop.ie/search.php?search_query=Parsons%2C%20Kelly%20%2F%20Doing%20Harm&section=product").
		Return(getHtmlNode(parsonsKellyDoingHarmTheBookshopSearch))

	Worker(validShelfURL, &websocket.Conn{})

	assert.Equal(t, len(db.GetAvailableBooks()), 1)
	assert.Equal(t, db.GetTotalBooksInAutomatedBookShelfCheck(), 1)
	assert.Equal(t, 1, len(db.GetKnownAuthors()))
}

func TestWorkerAddsOtherAuthorBooksWhenFlagIsEnabled(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	resetDBFields()
	db.SetAddMoreAuthorBooksToAvailableBooksList(true)

	mockController.On("Sleep", mock.Anything).After(100 * time.Nanosecond).Return()

	// Get the goodreads page
	mockController.On("GetPage", validShelfURL).Once().Return(getHtmlNode(stephenKingGoodreadsShelfOneBook))
	// Don't bother with validating websocket messages just yet
	mockController.On("WriteWsMessage", mock.Anything, mock.AnythingOfType(("*websocket.Conn")), mock.Anything).Return(nil)
	mockController.On("DeliverWebhook", mock.AnythingOfType("dtos.DiscordMsg")).Return(nil)

	mockController.On("GetPage", "https://thebookshop.ie/search.php?search_query=Parsons%2C%20Kelly%20%2F%20Doing%20Harm&section=product").
		Return(getHtmlNode(parsonsKellyDoingHarmTheBookshopSearch))

	Worker(validShelfURL, &websocket.Conn{})

	assert.Equal(t, len(db.GetAvailableBooks()), 2)
}

func TestWorkerCreatesBreadcrumbForCurrentCrawl(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	resetDBFields()

	mockController.On("Sleep", mock.Anything).After(100 * time.Nanosecond).Return()

	// Get the goodreads page
	mockController.On("GetPage", validShelfURL).Once().Return(getHtmlNode(stephenKingGoodreadsShelfOneBook))
	// Don't bother with validating websocket messages just yet
	mockController.On("WriteWsMessage", mock.Anything, mock.AnythingOfType(("*websocket.Conn")), mock.Anything).Return(nil)
	mockController.On("DeliverWebhook", mock.AnythingOfType("dtos.DiscordMsg")).Return(nil)

	mockController.On("GetPage", "https://thebookshop.ie/search.php?search_query=Parsons%2C%20Kelly%20%2F%20Doing%20Harm&section=product").
		Return(getHtmlNode(parsonsKellyDoingHarmTheBookshopSearch))

	Worker(validShelfURL, &websocket.Conn{})

	expectedCrawlBreadCrumb := dtos.RecentCrawlBreadcrumb{
		CrawlKey: "26367680-read",
		ShelfURL: validShelfURL,
	}
	assert.Equal(t, expectedCrawlBreadCrumb, db.GetRecentCrawlBreadcrumbs()[0])
}

func TestCheckAvailabilityOfExistingAvailableBooksListNoticesBooksThatAreNoLongerAvailable(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)
	resetDBFields()

	previouslyAvailableBooks := []dtos.AvailableBook{
		{
			BookInfo: dtos.BasicGoodReadsBook{
				Title:  "More Than This",
				Author: "Patrick Ness",
			},
			BookPurchaseInfo: dtos.TheBookshopBook{
				Title:  "More Than This",
				Author: "Patrick Ness",
			},
		},
	}
	db.SetAvailableBooks(previouslyAvailableBooks)
	db.SetSendAlertWhenBookNoLongerAvailable(true)

	// Return a search result where the previously available book can't be found
	// in as if its not available anymore since the last search
	mockController.On("GetPage", mock.AnythingOfType("string")).Return(getHtmlNode(parsonsKellyDoingHarmTheBookshopSearch))
	mockController.On("DeliverWebhook", mock.AnythingOfType("dtos.DiscordMsg")).Return(nil)

	assert.Equal(t, len(db.GetAvailableBooks()), 1)

	checkAvailabilityOfExistingAvailableBooksList()

	// Assert that the previously available book is removed
	// when it is found to not be available anymore
	assert.Equal(t, len(db.GetAvailableBooks()), 0)
	mockController.AssertNumberOfCalls(t, "DeliverWebhook", 1)
}

func TestFilterIgnoredAuthorsFiltersNothingWhenNoAuthorIsIgnored(t *testing.T) {
	resetDBFields()

	searchResult := dtos.EnchancedSearchResult{
		SearchBook: dtos.BasicGoodReadsBook{
			Title:  "Fallen Dragon",
			Author: "Patrick F. Hamilton",
		},
		TitleMatches: []dtos.TheBookshopBook{
			{
				Title: "Fallen Dragon",
			},
		},
		AuthorMatches: []dtos.TheBookshopBook{
			{
				Author: "Patrick F. Hamilton",
			},
		},
	}

	filteredSearchResults := filterIgnoredAuthors(searchResult)

	assert.Equal(t, len(searchResult.TitleMatches), len(filteredSearchResults.TitleMatches))
	assert.Equal(t, len(searchResult.AuthorMatches), len(filteredSearchResults.AuthorMatches))
}

func TestFilterIgnoredAuthorsFiltersOutIgnoredAuthors(t *testing.T) {
	resetDBFields()
	ignoredAuthor := "Patrick F. Hamilton"

	searchResult := dtos.EnchancedSearchResult{
		SearchBook: dtos.BasicGoodReadsBook{
			Title:  "Fallen Dragon",
			Author: ignoredAuthor,
		},
		TitleMatches: []dtos.TheBookshopBook{
			{
				Title:  "The Knife of Never Letting Go",
				Author: "Patrick Ness",
			},
		},
		AuthorMatches: []dtos.TheBookshopBook{
			{
				Author: ignoredAuthor,
			},
		},
	}

	db.AddAuthorToKnownAuthors(ignoredAuthor)
	db.ToggleAuthorIgnore(ignoredAuthor)
	filteredSearchResults := filterIgnoredAuthors(searchResult)

	assert.Equal(t, len(filteredSearchResults.TitleMatches), 1)
	assert.Equal(t, len(filteredSearchResults.AuthorMatches), 0)
}

func TestFindBooksThatAreNowNotAvailableReturnsBooksThatAreNoLongerAvailable(t *testing.T) {
	resetDBFields()
	wiseMansFear := dtos.AvailableBook{
		BookInfo: dtos.BasicGoodReadsBook{
			ID: "wiseMansFear",
		},
	}
	nameOfTheWind := dtos.AvailableBook{
		BookInfo: dtos.BasicGoodReadsBook{
			ID: "nameOfTheWind",
		},
	}
	booksThatWereAvailable := []dtos.AvailableBook{
		wiseMansFear,
		nameOfTheWind,
	}
	bookThatAreNowAvailable := []dtos.AvailableBook{
		nameOfTheWind,
	}

	booksThatAreNoLongerAvailable := findBooksThatAreNowNotAvailable(booksThatWereAvailable, bookThatAreNowAvailable)

	assert.Equal(t, len(booksThatAreNoLongerAvailable), 1)
	assert.Equal(t, wiseMansFear, booksThatAreNoLongerAvailable[0])
}

func TestFindBooksThatAreNowNotAvailableReturnsNothingWhenNewBooksAreAvailableNow(t *testing.T) {
	resetDBFields()
	wiseMansFear := dtos.AvailableBook{
		BookInfo: dtos.BasicGoodReadsBook{
			ID: "wiseMansFear",
		},
	}
	nameOfTheWind := dtos.AvailableBook{
		BookInfo: dtos.BasicGoodReadsBook{
			ID: "nameOfTheWind",
		},
	}
	booksThatWereAvailable := []dtos.AvailableBook{
		wiseMansFear,
	}
	bookThatAreNowAvailable := []dtos.AvailableBook{
		nameOfTheWind,
		wiseMansFear,
	}

	booksThatAreNoLongerAvailable := findBooksThatAreNowNotAvailable(booksThatWereAvailable, bookThatAreNowAvailable)

	assert.Equal(t, len(booksThatAreNoLongerAvailable), 0)
}

func TestFreeShippingNotificationNotTriggeredWhenTotalIsUnderThreshold(t *testing.T) {
	resetDBFields()
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)

	availableBooks := []dtos.AvailableBook{
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Price: "€18.50",
			},
		},
	}
	db.SetAvailableBooks(availableBooks)

	sendFreeShippingWebhookIfFreeShippingEligible()

	mockController.AssertNotCalled(t, "DeliverWebhook")
}

func TestFreeShippingNotificationIsTriggeredWhenTotalIsOverThreshold(t *testing.T) {
	resetDBFields()
	mockController := controller.MockCntrInterface{}
	controller.SetController(&mockController)

	availableBooks := []dtos.AvailableBook{
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Price: "€20.50",
			},
		},
	}
	db.SetAvailableBooks(availableBooks)
	mockController.On("DeliverWebhook", mock.AnythingOfType("dtos.DiscordMsg")).Return(nil)

	sendFreeShippingWebhookIfFreeShippingEligible()

	mockController.AssertNumberOfCalls(t, "DeliverWebhook", 1)
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
