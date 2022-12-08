package db

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

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

func nameTempRedisKeys() {
	AVAILABLE_BOOKS = "test-availableBooks"
	RECENT_CRAWL_BREADCRUMBS = "test-recentCrawls"
	AUTOMATED_BOOK_SHELF_CHECK_URL = "test-automatedBookShelfCheck"
	AUTOMATED_BOOK_SHELF_CRAWL_TIME = "test-automatedBookShelfCrawlTime"
	DISCORD_WEBHOOK_URL = "test-discordWebHookURL"
	DISCORD_MESSAGE_FORMAT = "test-discordMessageFormat"
	SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE = "test-sendAlertWhenBookNoLongerAvailable"
	SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN = "test-sendAlertWhenFreeShippingKicksIn"
}

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)

	connectToDevRedisDatabase()
	SetTestDataIdentifiers()
	DEFAULT_TTL = time.Duration(5 * time.Second)

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

	AddNewCrawlBreadcrumb(breadCrumbs[0].ShelfURL)
	assert.Len(t, GetRecentCrawlBreadcrumbs(), 1)

	AddNewCrawlBreadcrumb(breadCrumbs[1].ShelfURL)
	assert.Len(t, GetRecentCrawlBreadcrumbs(), 2)

	AddNewCrawlBreadcrumb(breadCrumbs[2].ShelfURL)
	assert.Len(t, GetRecentCrawlBreadcrumbs(), 2)
}

func TestBreadCrumbMakesCorrectCrawlKey(t *testing.T) {
	AddNewCrawlBreadcrumb(sharonFantasyLink)
	assert.Equal(t, "sharon-fantasy", GetRecentCrawlBreadcrumbs()[0].CrawlKey)
}

func TestBreadCrumbMakesCorrectCrawlKeyWhenNoUsernameGiven(t *testing.T) {
	AddNewCrawlBreadcrumb("https://www.goodreads.com/review/list/26367680?shelf=read")
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

func resetDBFields() {
	SetKnownAuthors([]dtos.KnownAuthor{})
	SetAddMoreAuthorBooksToAvailableBooksList(false)
	SetAvailableBooks([]dtos.AvailableBook{})
	SetSendAlertOnlyWhenFreeShippingKicksIn(false)
	SetSendAlertWhenBookNoLongerAvailable(false)
	SetDiscordMessageFormat("small")
	SetRecentCrawlBreadcrumbs([]dtos.RecentCrawlBreadcrumb{})
}
