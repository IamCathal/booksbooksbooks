package thebookshop

import (
	"log"
	"os"
	"testing"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"go.uber.org/zap"
	"gotest.tools/assert"
)

var (
	validShelfURL = "https://www.goodreads.com/review/list/26367680?shelf=read"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)

	code := m.Run()

	os.Exit(code)
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
