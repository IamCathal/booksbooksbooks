package db

import (
	"log"
	"os"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func connectToDevRedisDatabase() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
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
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)

	connectToDevRedisDatabase()
	nameTempRedisKeys()
	DEFAULT_TTL = time.Duration(5 * time.Second)

	code := m.Run()

	os.Exit(code)
}

func TestGetAvailableBooksReturnsAnEmptyArrayWhenNothingIsSet(t *testing.T) {
	assert.Empty(t, GetAvailableBooks())
}

func TestGetAvailableBooksReturnsSomething(t *testing.T) {
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

func TestGetDiscordMessageFormatReturnsSmallWhenNotSet(t *testing.T) {
	assert.Equal(t, GetDiscordMessageFormat(), "small")
}

func TestGetSendAlertWhenBookIsNoLongerAvailableReturnsFalseWhenNotSet(t *testing.T) {
	assert.False(t, GetSendAlertWhenBookNoLongerAvailable())
}

func TestGetSendAlertOnlyWhenFreeShippingKicksInReturnsFalseWhenNotSet(t *testing.T) {
	assert.False(t, GetSendAlertOnlyWhenFreeShippingKicksIn())
}
