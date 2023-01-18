package db

import (
	"fmt"
	"log"
	"os"
	"testing"

	redis "github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var (
	sharonFantasyLink        = "https://www.goodreads.com/review/list/1753152-sharon?shelf=fantasy"
	kingCurrentlyReadingLink = "https://www.goodreads.com/review/list/26367680?shelf=currently-reading"
)

func connectToDevRedisDatabase() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	response, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to redis. Response: '%s' error: '%s'", response, err)
	}
	redisClient = rdb
}

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	// c.OutputPaths = []string{"/dev/null"}
	c.OutputPaths = []string{"stdout"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)

	connectToDevRedisDatabase()
	SetTestDataIdentifiers()
	resetDBFields()

	code := m.Run()

	os.Exit(code)
}

func TestGetAvailableBooksReturnsAnEmptyArrayWhenNothingIsSet(t *testing.T) {
	resetDBFields()
	assert.Empty(t, GetAvailableBooks())
}

func TestGetAvailableBooksReturnsSomething(t *testing.T) {
	resetDBFields()
	availableBooks := []dtos.AvailableBook{
		{
			BookInfo: dtos.BasicGoodReadsBook{
				Title: "a book",
			},
		},
	}
	SetAvailableBooks(availableBooks)

	assert.ElementsMatch(t, GetAvailableBooks(), availableBooks)
}

func TestResetAvailableBooksClearsAllAvailableBooks(t *testing.T) {
	resetDBFields()
	availableBooks := []dtos.AvailableBook{
		{
			BookInfo: dtos.BasicGoodReadsBook{
				Title: "a book",
			},
		},
	}
	SetAvailableBooks(availableBooks)

	ResetAvailableBooks()

	assert.Empty(t, GetAvailableBooks())
}

func TestNoDuplicateBreadcrumbsAreSaved(t *testing.T) {
	resetDBFields()
	breadCrumbs := []dtos.RecentCrawlBreadcrumb{
		{
			ShelfURL: "https://www.goodreads.com/review/list/26367680?shelf=read",
		},
		{
			ShelfURL: "https://www.goodreads.com/review/list/26367680?shelf=currently-reading",
		},
		{
			ShelfURL: "https://www.goodreads.com/review/list/26367680?shelf=read",
		},
	}
	assert.Empty(t, GetRecentCrawlBreadcrumbs())

	AddNewCrawlBreadcrumb(breadCrumbs[0].ShelfURL, 0)
	assert.Len(t, GetRecentCrawlBreadcrumbs(), 1)

	AddNewCrawlBreadcrumb(breadCrumbs[1].ShelfURL, 0)
	assert.Len(t, GetRecentCrawlBreadcrumbs(), 2)

	AddNewCrawlBreadcrumb(breadCrumbs[2].ShelfURL, 0)
	assert.Len(t, GetRecentCrawlBreadcrumbs(), 2)
}

func TestBreadCrumbMakesCorrectCrawlKey(t *testing.T) {
	AddNewCrawlBreadcrumb(sharonFantasyLink, 0)
	assert.Equal(t, "sharon-fantasy", GetRecentCrawlBreadcrumbs()[0].CrawlKey)
}

func TestBreadCrumbMakesCorrectCrawlKeyWhenNoUsernameGiven(t *testing.T) {
	AddNewCrawlBreadcrumb("https://www.goodreads.com/review/list/26367680?shelf=read", 0)
	assert.Equal(t, "26367680-read", GetRecentCrawlBreadcrumbs()[0].CrawlKey)
}

func TestGetDiscordMessageFormatReturnsSmallWhenNotSet(t *testing.T) {
	resetDBFields()
	assert.Equal(t, GetDiscordMessageFormat(), "small")
}

func TestGetSendAlertWhenBookIsNoLongerAvailableReturnsFalseWhenNotSet(t *testing.T) {
	resetDBFields()
	assert.False(t, GetSendAlertWhenBookNoLongerAvailable())
}

func TestGetSendAlertOnlyWhenFreeShippingKicksInReturnsFalseWhenNotSet(t *testing.T) {
	resetDBFields()
	assert.False(t, GetSendAlertOnlyWhenFreeShippingKicksIn())
}

func TestGetAvailableBooksMap(t *testing.T) {
	resetDBFields()
	duplicateLink := "duplicateLink"
	availableBooks := []dtos.AvailableBook{
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link: "https://cathaloc.dev",
			},
		},
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link: duplicateLink,
			},
		},
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link: duplicateLink,
			},
		},
	}
	SetAvailableBooks(availableBooks)

	assert.Len(t, GetAvailableBooksMap(), 2)
}

func TestGetKeyForRecentCrawlBreadCrumb(t *testing.T) {
	assert.Equal(t, "sharon-fantasy", GetKeyForRecentCrawlBreadcrumb(sharonFantasyLink))
}

func TestGetKeyForRecentCrawlBreadCrumbsHandlesUrlsWithoutUsernames(t *testing.T) {
	assert.Equal(t, "26367680-currently-reading", GetKeyForRecentCrawlBreadcrumb(kingCurrentlyReadingLink))
}

func TestRemoveDuplicateAvailableBooks(t *testing.T) {
	resetDBFields()
	duplicateLink := "duplicate link"
	availableBooks := []dtos.AvailableBook{
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link: duplicateLink,
			},
		},
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link: duplicateLink,
			},
		},
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link: "a new link",
			},
		},
	}

	assert.Len(t, removeDuplicateAvailableBooks(availableBooks), 2)
}

func TestStrToBool(t *testing.T) {
	assert.True(t, strToBool("true"))
	assert.False(t, strToBool("false"))
}

func TestIgnoreAuthorFlow(t *testing.T) {
	resetDBFields()
	initialKnownAuthors := []dtos.KnownAuthor{
		{
			Name:   "Patrick Rothfuss",
			Ignore: false,
		},
		{
			Name:   "Ken MacLeod",
			Ignore: false,
		},
	}
	SetKnownAuthors(initialKnownAuthors)
	assert.ElementsMatch(t, initialKnownAuthors, GetKnownAuthors())

	// Toggle ignore on an author
	initialKnownAuthors[1].Ignore = true
	ToggleAuthorIgnore(initialKnownAuthors[1].Name)
	assert.ElementsMatch(t, initialKnownAuthors, GetKnownAuthors())

	// try to add an already existing author, no change to the existing entry
	// and ignore status does not change
	AddAuthorToKnownAuthors("Ken MacLeod")
	assert.ElementsMatch(t, initialKnownAuthors, GetKnownAuthors())
}

func TestPurgeAuthorFromAvailableBooksPurgesSingleIgnoredAuthor(t *testing.T) {
	resetDBFields()
	retroActivelyPurgeAuthor := "Ken Mc Leod"
	availableBooks := []dtos.AvailableBook{
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link:   "https://cathaloc.dev",
				Author: retroActivelyPurgeAuthor,
			},
		},
	}
	SetAvailableBooks(availableBooks)
	assert.Equal(t, GetAvailableBooks(), availableBooks)

	AddAuthorToKnownAuthors(retroActivelyPurgeAuthor)
	ToggleAuthorIgnore(retroActivelyPurgeAuthor)

	PurgeIgnoredAuthorsFromAvailableBooks()

	// Expect available books left from the retroactively purged author
	assert.Empty(t, GetAvailableBooks())
}

func TestPurgeAuthorFromAvailableBooksPurgesIgnoredAuthorsAndLeavesTheRest(t *testing.T) {
	resetDBFields()
	retroActivelyPurgeAuthor := "Ken Mc Leod"
	normalAuthor := "Stephen Lawhead"
	normalAuthor2 := fmt.Sprintf("%s 2", retroActivelyPurgeAuthor)

	availableBooks := []dtos.AvailableBook{
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link:   "https://cathaloc.dev",
				Author: retroActivelyPurgeAuthor,
			},
		},
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link:   "https://cathaloc.dev/fyp",
				Author: normalAuthor,
			},
		},
		{
			BookPurchaseInfo: dtos.TheBookshopBook{
				Link:   "https://cathaloc.dev/fyp",
				Author: normalAuthor2,
			},
		},
	}
	SetAvailableBooks(availableBooks)
	assert.Equal(t, GetAvailableBooks(), availableBooks)

	AddAuthorToKnownAuthors(retroActivelyPurgeAuthor)
	ToggleAuthorIgnore(retroActivelyPurgeAuthor)
	AddAuthorToKnownAuthors(normalAuthor)
	AddAuthorToKnownAuthors(normalAuthor2)

	PurgeIgnoredAuthorsFromAvailableBooks()

	// Expect available books left from the retroactively purged author
	assert.Equal(t, len(GetAvailableBooks()), 2)
}

func TestIgnoreBook(t *testing.T) {
	resetDBFields()
	ignoredBook := dtos.AvailableBook{
		BookInfo: dtos.BasicGoodReadsBook{
			Title: "Century Rain",
		},
		BookPurchaseInfo: dtos.TheBookshopBook{
			Link: "testLink",
		},
		Ignore: false,
	}
	AddAvailableBook(ignoredBook)

	IgnoreBook(ignoredBook.BookPurchaseInfo.Link)

	allAvailableBooks := GetAvailableBooks()

	assert.Equal(t, 1, len(allAvailableBooks))
	assert.Equal(t, true, allAvailableBooks[0].Ignore)
}

func TestUnignoreBook(t *testing.T) {
	resetDBFields()
	ignoredBook := dtos.AvailableBook{
		BookInfo: dtos.BasicGoodReadsBook{
			Title: "Century Rain",
		},
		BookPurchaseInfo: dtos.TheBookshopBook{
			Link: "testLink",
		},
		Ignore: true,
	}
	AddAvailableBook(ignoredBook)

	UnignoreBook(ignoredBook.BookPurchaseInfo.Link)

	allAvailableBooks := GetAvailableBooks()

	assert.Equal(t, 1, len(allAvailableBooks))
	assert.Equal(t, false, allAvailableBooks[0].Ignore)
}

func TestGetIgnoredAndNonIgnoredCountOfAvailableBooks(t *testing.T) {
	resetDBFields()
	ignoredBook := dtos.AvailableBook{
		BookInfo: dtos.BasicGoodReadsBook{
			Title: "Century Rain",
		},
		BookPurchaseInfo: dtos.TheBookshopBook{
			Link: "testLink",
		},
		Ignore: false,
	}
	AddAvailableBook(ignoredBook)

	actualNonIgnoredCount, actualIgnoredCount := GetIgnoredAndNonIgnoredCountOfAvailableBooks()

	assert.Equal(t, 1, actualNonIgnoredCount)
	assert.Equal(t, 0, actualIgnoredCount)
}

func TestIgnoredAuthor(t *testing.T) {
	resetDBFields()
	knownAuthorsList := []dtos.KnownAuthor{
		{
			Name:   "Iain Banks",
			Ignore: true,
		},
	}
	SetKnownAuthors(knownAuthorsList)

	assert.True(t, IsIgnoredAuthor(knownAuthorsList[0].Name))
}

func TestIgnoredAuthorReturnsFalseForAnUnknownAuthor(t *testing.T) {
	resetDBFields()
	knownAuthorsList := []dtos.KnownAuthor{
		{
			Name:   "Iain Banks",
			Ignore: true,
		},
	}
	SetKnownAuthors(knownAuthorsList)

	assert.False(t, IsIgnoredAuthor("Mac Leod"))
}

func TestAddAuthorToKnownAuthorDoesntAddReverseOrderAuthorDuplicate(t *testing.T) {
	// if "Suzanne Collins" is a known author then don't add
	// "Collins, Suzanne" as if she's new
	resetDBFields()

	AddAuthorToKnownAuthors("Suzanne Collins")
	assert.Equal(t, 1, len(GetKnownAuthors()))

	AddAuthorToKnownAuthors("Collins, Suzanne")
	assert.Equal(t, 1, len(GetKnownAuthors()))
}

func TestRemoveDuplicateAuthorsDisregardingReverseOrder(t *testing.T) {
	inputAuthorsList := []dtos.KnownAuthor{
		{
			Name:   "Suzanne Collins",
			Ignore: false,
		},
		{
			Name:   "Collins, Suzanne",
			Ignore: false,
		},
	}
	expectedFilteredAuthorsList := []dtos.KnownAuthor{
		{
			Name:   "Suzanne Collins",
			Ignore: false,
		},
	}

	actualFilteredAuthorsList := removeDuplicateAuthorsDisregardingReverseOrder(inputAuthorsList)

	assert.Equal(t, expectedFilteredAuthorsList, actualFilteredAuthorsList)
}

func TestGetAuthorNameTokens(t *testing.T) {
	assert.Equal(t, []string{"Collins", "Suzanne"}, getAuthorNameTokens("Collins, Suzanne"))
}

func TestGetAuthorNameTokensWithAuthorThatDoesntHaveComma(t *testing.T) {
	assert.Equal(t, []string{"Suzanne", "Collins"}, getAuthorNameTokens("Suzanne Collins"))
}

func TestConvertAuthorNameToSortedString(t *testing.T) {
	assert.Equal(t, "CollinsSuzanne", convertAuthorNameToSortedString("Collins, Suzanne"))
}

func TestConvertAuthorNameToSortedStringWithLotsOfUnnecessaryCharacters(t *testing.T) {
	assert.Equal(t, "CollinsSuzanne", convertAuthorNameToSortedString("Collins,   ------Suzanne"))
}

func TestAddShelfToShelvesToCrawlDoesNotAddDuplicates(t *testing.T) {
	shelf := "https://www.goodreads.com/review/list/26367680?shelf=read"
	firstShelfToCrawl := dtos.ShelfToCrawl{
		CrawlKey:  GetKeyForRecentCrawlBreadcrumb(shelf),
		ShelfURL:  shelf,
		BookCount: 12,
	}

	AddShelfToShelvesToCrawl(firstShelfToCrawl)
	assert.Equal(t, 1, len(GetShelvesToCrawl()))

	AddShelfToShelvesToCrawl(firstShelfToCrawl)
	assert.Equal(t, 1, len(GetShelvesToCrawl()))
}

func resetDBFields() {
	SetKnownAuthors([]dtos.KnownAuthor{})
	setShelvesToCrawl([]dtos.ShelfToCrawl{})
	SetAddMoreAuthorBooksToAvailableBooksList(false)
	SetAvailableBooks([]dtos.AvailableBook{})
	SetSendAlertOnlyWhenFreeShippingKicksIn(false)
	SetSendAlertWhenBookNoLongerAvailable(false)
	SetDiscordMessageFormat("small")
	SetRecentCrawlBreadcrumbs([]dtos.RecentCrawlBreadcrumb{})
}
