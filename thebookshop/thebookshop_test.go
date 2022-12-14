package thebookshop

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

var (
	validShelfURL                          = "https://www.goodreads.com/review/list/26367680?shelf=read"
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
	search.SetLogger(logger)
	db.SetLogger(logger)
	loadMockSearchResults()

	db.ConnectToRedis()
	db.SetTestDataIdentifiers()

	code := m.Run()

	os.Exit(code)
}

func loadMockSearchResults() {
	data, err := os.ReadFile("../testData/parsonsKellyDoingHarmTheBookshop.html")
	if err != nil {
		panic(err)
	}
	parsonsKellyDoingHarmTheBookshopSearch = string(data)
}

func TestURLEncodeBookSearch(t *testing.T) {
	actualEncodedURIParams := "search_query=TOLKIEN%20%2F%20THE%20LORD%20OF%20THE%20RINGS&section=product"
	bookInfo := dtos.BasicGoodReadsBook{
		Author: "TOLKIEN",
		Title:  "THE LORD OF THE RINGS",
	}
	assert.Equal(t, actualEncodedURIParams, urlEncodeBookSearch(bookInfo))
}

func TestSearchForBookRespectsSleepDurationBetweenRequests(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	controller.SetController(mockController)

	bookSearchResultsChan := make(chan dtos.EnchancedSearchResult, 200)
	mockController.On("GetPage", mock.AnythingOfType("string")).Return(getHtmlNode(parsonsKellyDoingHarmTheBookshopSearch))

	startTime := time.Now()
	SLEEP_DURATION = time.Duration(12 * time.Millisecond)
	timesCalled := 9

	for i := 0; i < timesCalled; i++ {
		SearchForBook(dtos.BasicGoodReadsBook{}, bookSearchResultsChan)
	}

	timeTaken := time.Since(startTime)

	perfectWorldTimeTaken := time.Duration(timesCalled) * SLEEP_DURATION
	expectedTimeTaken := perfectWorldTimeTaken.Abs().Milliseconds()

	// Allow for 10ms expected difference since sleep is not constant
	assert.Greater(t, timeTaken.Abs().Milliseconds(), expectedTimeTaken-10)
}

func getHtmlNode(webpageStr string) *html.Node {
	htmlNodeResponse, err := html.Parse(strings.NewReader(webpageStr))
	if err != nil {
		panic(err)
	}
	return htmlNodeResponse
}
