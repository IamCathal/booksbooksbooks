package integration

import (
	"log"
	"os"
	"testing"

	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"stdout"}
	globalLogFields := make(map[string]interface{})
	globalLogFields["service"] = "booksbooksbooks"
	c.InitialFields = globalLogFields
	testLogger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}

	logger = testLogger
	db.SetLogger(testLogger)
	db.ConnectToRedis()
	db.SetTestDataIdentifiers()

	cnt := controller.Cntr{}
	controller.SetController(cnt)

	code := m.Run()

	os.Exit(code)
}

func TestSearchGoodReadsForChasmCityFromTheBookShop(t *testing.T) {
	bookFromTheBookShop := dtos.TheBookshopBook{
		Title:  "Chasm City",
		Author: "Alastair Reynolds",
	}

	found, bookSearchResult := goodreads.SearchGoodreads(bookFromTheBookShop)

	expectedChasmCityBook := dtos.BasicGoodReadsBook{
		Title:      "Chasm City",
		Author:     "Alastair Reynolds",
		SeriesText: "",
	}
	assert.True(t, found, "expect to find a match")
	assert.Equal(t, expectedChasmCityBook.Title, bookSearchResult.Title)
	assert.Equal(t, expectedChasmCityBook.Author, bookSearchResult.Author)
	assert.Equal(t, expectedChasmCityBook.SeriesText, bookSearchResult.SeriesText)
}
