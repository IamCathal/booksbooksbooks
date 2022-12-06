package thebookshop

import (
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

var (
	validShelfURL        = "https://www.goodreads.com/review/list/26367680?shelf=read"
	rothfussSearchResult = ""
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
	loadMockSearchResults()

	code := m.Run()

	os.Exit(code)
}

func loadMockSearchResults() {
	data, err := os.ReadFile("testData/rothfussTheBookshopSearch.html")
	if err != nil {
		panic(err)
	}
	rothfussSearchResult = string(data)
}

func TestURLEncodeBookSearch(t *testing.T) {
	actualEncodedURIParams := "search_query=TOLKIEN%20%2F%20THE%20LORD%20OF%20THE%20RINGS&section=product"
	bookInfo := dtos.BasicGoodReadsBook{
		Author: "TOLKIEN",
		Title:  "THE LORD OF THE RINGS",
	}
	assert.Equal(t, actualEncodedURIParams, urlEncodeBookSearch(bookInfo))
}

func TestExtractAuthorFromTitleSplitBySlash(t *testing.T) {
	rawTitleText := "Tolkien, J. R. R. / The Lord of the Rings"
	expectedAuthor := "Tolkien, J. R. R."
	expectedTitle := "The Lord of the Rings"

	author, title := extractAuthorFromTitle(rawTitleText)

	assert.Equal(t, expectedAuthor, author)
	assert.Equal(t, expectedTitle, title)
}

func TestExtractAuthorFromTitleSplitByHyphen(t *testing.T) {
	rawTitleText := "Tolkien, J. R. R. - The Lord of the Rings"
	expectedAuthor := "Tolkien, J. R. R."
	expectedTitle := "The Lord of the Rings"

	author, title := extractAuthorFromTitle(rawTitleText)

	assert.Equal(t, expectedAuthor, author)
	assert.Equal(t, expectedTitle, title)
}

func TestExtractAuthorFromTitleReturnsEverythingWhenCantSplit(t *testing.T) {
	rawTitleText := "Tolkien, J. R. R. [] The Lord of the Rings"
	expectedAuthor := rawTitleText
	expectedTitle := rawTitleText

	author, title := extractAuthorFromTitle(rawTitleText)

	assert.Equal(t, expectedAuthor, author)
	assert.Equal(t, expectedTitle, title)
}

func TestSearchForBookRespectsSleepDurationBetweenRequests(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	controller.SetController(mockController)

	bookSearchResultsChan := make(chan dtos.EnchancedSearchResult, 200)
	mockController.On("GetPage", mock.AnythingOfType("string")).Return(getHtmlNode(rothfussSearchResult))

	startTime := time.Now()
	SLEEP_DURATION = time.Duration(12 * time.Millisecond)
	timesCalled := 9

	for i := 0; i < timesCalled; i++ {
		SearchForBook(dtos.BasicGoodReadsBook{}, bookSearchResultsChan)
	}

	timeTaken := time.Since(startTime)

	perfectWorldTimeTaken := time.Duration(timesCalled) * SLEEP_DURATION
	expectedTimeTaken := perfectWorldTimeTaken.Abs().Milliseconds()

	// Allow for 5ms expected difference since sleep is not constant
	assert.Greater(t, timeTaken.Abs().Milliseconds(), expectedTimeTaken-5)
}

func getHtmlNode(webpageStr string) *html.Node {
	htmlNodeResponse, err := html.Parse(strings.NewReader(webpageStr))
	if err != nil {
		panic(err)
	}
	return htmlNodeResponse
}
