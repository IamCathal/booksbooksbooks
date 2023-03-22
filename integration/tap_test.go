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
	logger                  *zap.Logger
	CHAOS_WALKING_SHELF     = "https://www.goodreads.com/review/list/164034456?shelf=chaoswalking"
	LORD_OF_THE_RINGS_SHELF = "https://www.goodreads.com/review/list/164034456?shelf=lordoftherings"
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
	goodreads.SetLogger(testLogger)
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

func TestGetsBooksFromShelf(t *testing.T) {
	expectedBooksFromShelf := []dtos.BasicGoodReadsBook{
		{
			Title:      "The Knife of Never Letting Go",
			Author:     "Ness, Patrick",
			SeriesText: "Chaos Walking, #1",
		},
		{
			Title:      "The Ask and the Answer",
			Author:     "Ness, Patrick",
			SeriesText: "Chaos Walking, #2",
		},
		{
			Title:      "Monsters of Men",
			Author:     "Ness, Patrick",
			SeriesText: "Chaos Walking, #3",
		},
	}
	shelfStatsChan := make(chan int, 5)
	booksFromShelfChan := make(chan dtos.BasicGoodReadsBook, 50)
	defer close(shelfStatsChan)
	defer close(booksFromShelfChan)

	actualBooksFromShelves := goodreads.GetBooksFromShelf(CHAOS_WALKING_SHELF, shelfStatsChan, booksFromShelfChan)

	assert.ObjectsAreEqualValues(expectedBooksFromShelf, actualBooksFromShelves)
}

func TestGetsBooksFromShelves(t *testing.T) {
	expectedBooksFromShelf := []dtos.BasicGoodReadsBook{
		{
			Title:      "The Knife of Never Letting Go",
			Author:     "Ness, Patrick",
			SeriesText: "Chaos Walking, #1",
		},
		{
			Title:      "The Ask and the Answer",
			Author:     "Ness, Patrick",
			SeriesText: "Chaos Walking, #2",
		},
		{
			Title:      "Monsters of Men",
			Author:     "Ness, Patrick",
			SeriesText: "Chaos Walking, #3",
		},
		{
			Title:      "The Hobbit",
			Author:     "Tolkien, J.R.R.",
			SeriesText: "The Lord Of The Rings, #0",
		},
	}
	shelfStatsChan := make(chan int, 5)
	booksFromShelfChan := make(chan dtos.BasicGoodReadsBook, 50)
	defer close(booksFromShelfChan)

	actualBooksFromShelves := goodreads.GetBooksFromShelves([]string{CHAOS_WALKING_SHELF, LORD_OF_THE_RINGS_SHELF}, shelfStatsChan, booksFromShelfChan)

	assert.ObjectsAreEqualValues(expectedBooksFromShelf, actualBooksFromShelves)
}
