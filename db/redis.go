package db

import (
	"context"
	"encoding/json"
	"os"
	"time"

	redis "github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"go.uber.org/zap"
)

var (
	logger      *zap.Logger
	ctx         = context.Background()
	redisClient *redis.Client

	AVAILABLE_BOOKS                             = "availableBooks"
	RECENT_CRAWL_BREADCRUMBS                    = "recentCrawls"
	AUTOMATED_BOOK_SHELF_CHECK_URL              = "automatedBookShelfCheck"
	AUTOMATED_BOOK_SHELF_CRAWL_TIME             = "automatedBookShelfCrawlTime"
	DISCORD_WEBHOOK_URL                         = "discordWebHookURL"
	DISCORD_MESSAGE_FORMAT                      = "discordMessageFormat"
	SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE    = "sendAlertWhenBookNoLongerAvailable"
	SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN = "sendAlertWhenFreeShippingKicksIn"
	DEFAULT_TTL                                 = time.Duration(0)
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func ConnectToRedis() {
	logger.Info("Connecting to redis...", zap.String("diagnostics", "redis"))

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: "",
		DB:       0,
	})
	redisClient = rdb
	response, err := redisClient.Ping(ctx).Result()
	if err != nil {
		logger.Sugar().Fatalf("Could not connect to redis. Response: '%s' error: '%s'", response, err)
	}

	logger.Info("Redis connection successfully initialised")
}

func ResetAvailableBooks() {
	err := redisClient.Set(ctx, AVAILABLE_BOOKS, []byte(""), DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func AddAvailableBook(newBook dtos.AvailableBook) {
	availableBooks := GetAvailableBooks()
	newBook.Ignore = false
	availableBooks = append(availableBooks, newBook)

	availableBooksWithNoDuplicates := removeDuplicateAvailableBooks(availableBooks)
	jsonAvailableBooks, err := json.Marshal(availableBooksWithNoDuplicates)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	err = redisClient.Set(ctx, AVAILABLE_BOOKS, jsonAvailableBooks, time.Duration(DEFAULT_TTL)).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func SetAvailableBooks(availableBooks []dtos.AvailableBook) {
	availableBooksJson, err := json.Marshal(availableBooks)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, AVAILABLE_BOOKS, availableBooksJson, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAvailableBooks() []dtos.AvailableBook {
	availableBooksStr, err := redisClient.Get(ctx, AVAILABLE_BOOKS).Result()
	if err == redis.Nil {
		return []dtos.AvailableBook{}
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	availableBooks := []dtos.AvailableBook{}
	if availableBooksStr != "" {
		err = json.Unmarshal([]byte(availableBooksStr), &availableBooks)
		if err != nil {
			logger.Sugar().Fatal(err)
		}
	}
	return availableBooks
}

func GetRecentCrawlBreadcrumbs() []dtos.RecentCrawlBreadcrumb {
	recentCrawlBreadcrumbs, err := redisClient.Get(ctx, RECENT_CRAWL_BREADCRUMBS).Result()
	if err == redis.Nil {
		return []dtos.RecentCrawlBreadcrumb{}
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	recentCrawlBreadcrumbsArr := []dtos.RecentCrawlBreadcrumb{}
	if recentCrawlBreadcrumbs != "" {
		err = json.Unmarshal([]byte(recentCrawlBreadcrumbs), &recentCrawlBreadcrumbsArr)
		if err != nil {
			logger.Sugar().Fatal(err)
		}
	}
	return recentCrawlBreadcrumbsArr
}

func AddNewCrawlBreadcrumb(shelfURL string) {
	recentCrawls := GetRecentCrawlBreadcrumbs()

	updatedCrawlBreadcrumbs := []dtos.RecentCrawlBreadcrumb{
		{
			CrawlKey: getKeyForRecentCrawlBreadcrumb(shelfURL),
			ShelfURL: shelfURL,
		},
	}
	logger.Sugar().Infof("Creating new crawl breadcrumb with key: %s for shelfURL: %s",
		updatedCrawlBreadcrumbs[0].CrawlKey, updatedCrawlBreadcrumbs[0].ShelfURL)

	updatedCrawlBreadcrumbs = append(updatedCrawlBreadcrumbs, recentCrawls...)
	updatedCrawlBreadcrumbs = removeDuplicateRecentCrawls(updatedCrawlBreadcrumbs)

	jsonCrawlBreadcrumbs, err := json.Marshal(updatedCrawlBreadcrumbs)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, RECENT_CRAWL_BREADCRUMBS, jsonCrawlBreadcrumbs, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func SetAutomatedBookShelfCheck(shelfURL string) {
	err := redisClient.Set(ctx, AUTOMATED_BOOK_SHELF_CHECK_URL, shelfURL, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAutomatedBookShelfCheck() string {
	shelfURL, err := redisClient.Get(ctx, AUTOMATED_BOOK_SHELF_CHECK_URL).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return shelfURL
}

func SetDiscordWebhookURL(webhookURL string) {
	err := redisClient.Set(ctx, DISCORD_WEBHOOK_URL, webhookURL, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetDiscordWebhookURL() string {
	webhookURL, err := redisClient.Get(ctx, DISCORD_WEBHOOK_URL).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return webhookURL
}

func SetDiscordMessageFormat(format string) {
	err := redisClient.Set(ctx, DISCORD_MESSAGE_FORMAT, format, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetDiscordMessageFormat() string {
	format, err := redisClient.Get(ctx, DISCORD_MESSAGE_FORMAT).Result()
	if err == redis.Nil {
		SetDiscordMessageFormat("small")
		return GetDiscordMessageFormat()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if format == "" {
		SetDiscordMessageFormat("small")
		return GetDiscordMessageFormat()
	}
	return format
}

func SetAutomatedBookShelfCrawlTime(time string) {
	err := redisClient.Set(ctx, AUTOMATED_BOOK_SHELF_CRAWL_TIME, time, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAutomatedBookShelfCrawlTime() string {
	time, err := redisClient.Get(ctx, AUTOMATED_BOOK_SHELF_CRAWL_TIME).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return time
}

func SetSendAlertWhenBookNoLongerAvailable(enabled bool) {
	err := redisClient.Set(ctx, SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetSendAlertWhenBookNoLongerAvailable() bool {
	enabled, err := redisClient.Get(ctx, SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE).Result()
	if err == redis.Nil {
		SetSendAlertWhenBookNoLongerAvailable(false)
		return GetSendAlertWhenBookNoLongerAvailable()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetSendAlertWhenBookNoLongerAvailable(false)
		return GetSendAlertWhenBookNoLongerAvailable()
	}
	return strToBool(enabled)
}

func SetSendAlertOnlyWhenFreeShippingKicksIn(enabled bool) {
	err := redisClient.Set(ctx, SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetSendAlertOnlyWhenFreeShippingKicksIn() bool {
	enabled, err := redisClient.Get(ctx, SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN).Result()
	if err == redis.Nil {
		SetSendAlertOnlyWhenFreeShippingKicksIn(false)
		return GetSendAlertOnlyWhenFreeShippingKicksIn()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetSendAlertOnlyWhenFreeShippingKicksIn(false)
		return GetSendAlertOnlyWhenFreeShippingKicksIn()
	}
	return strToBool(enabled)
}
