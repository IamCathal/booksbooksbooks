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
	mock "github.com/stretchr/testify/mock"
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
	goodreads.SetController(&mockController)
	thebookshop.SetController(&mockController)
	SetController(&mockController)

	mockController.On("Sleep", mock.Anything).After(1 * time.Millisecond).Return()

	// Get the goodreads page
	mockController.On("GetPage", validShelfURL).Once().Return(getHtmlNode(stephenKingGoodreadsShelfOneBook))
	// Don't bother with validating websocket messages just yet
	mockController.On("WriteWsMessage", mock.Anything, mock.Anything).Return(nil)

	mockController.On("GetPage", "https://thebookshop.ie/search.php?search_query=Parsons%2C%20Kelly%20%2F%20Doing%20Harm&section=product").
		Return(getHtmlNode(parsonsKellyDoingHarmTheBookshopSearch))

	Worker(validShelfURL, &websocket.Conn{})

	expectedCrawlBreadCrumb := dtos.RecentCrawlBreadcrumb{
		CrawlKey: "26367680-read",
		ShelfURL: validShelfURL,
	}
	assert.Equal(t, expectedCrawlBreadCrumb, db.GetRecentCrawlBreadcrumbs()[0])

	assert.Equal(t, len(db.GetAvailableBooks()), 2)
	assert.Equal(t, db.GetTotalBooksInAutomatedBookShelfCheck(), 1)
	assert.Equal(t, 1, len(db.GetKnownAuthors()))
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