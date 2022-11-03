package db

import (
	"context"
	"encoding/json"
	"os"

	redis "github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"go.uber.org/zap"
)

var (
	logger      *zap.Logger
	ctx         = context.Background()
	redisClient *redis.Client

	AVAILABLE_BOOKS            = "availableBooks"
	RECENT_CRAWLS              = "recentCrawls"
	AUTOMATED_BOOK_SHELF_CHECK = "automatedBookShelfCheck"
	DISCORD_WEBHOOK_URL        = "discordWebHookURL"
	DISCORD_MESSAGE_FORMAT     = "discordMessageFormat"
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
	err := redisClient.Set(ctx, AVAILABLE_BOOKS, []byte(""), 0).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func AddAvailableBook(newBook dtos.AvailableBook) {
	availableBooks := GetAvailableBooks()
	availableBooks = append(availableBooks, newBook)

	availableBooksWithNoDuplicates := removeDuplicateAvailableBooks(availableBooks)
	jsonAvailableBooks, err := json.Marshal(availableBooksWithNoDuplicates)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	err = redisClient.Set(ctx, AVAILABLE_BOOKS, jsonAvailableBooks, 0).Err()
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

func GetRecentCrawls() []dtos.RecentCrawl {
	recentCrawls, err := redisClient.Get(ctx, RECENT_CRAWLS).Result()
	if err == redis.Nil {
		return []dtos.RecentCrawl{}
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	recentCrawlsArr := []dtos.RecentCrawl{}
	if recentCrawls != "" {
		err = json.Unmarshal([]byte(recentCrawls), &recentCrawlsArr)
		if err != nil {
			logger.Sugar().Fatal(err)
		}
	}
	return removeDuplicateRecentCrawls(recentCrawlsArr)
}

func SaveRecentCrawlStats(shelfURL string) {
	recentCrawls := GetRecentCrawls()

	newRecentCrawl := []dtos.RecentCrawl{
		{
			CrawlKey: getKeyForRecentCrawl(shelfURL),
			ShelfURL: shelfURL,
		},
	}
	newRecentCrawl = append(newRecentCrawl, recentCrawls...)
	newRecentCrawl = removeDuplicateRecentCrawls(newRecentCrawl)

	jsonCrawls, err := json.Marshal(newRecentCrawl)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, RECENT_CRAWLS, jsonCrawls, 0).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func SetAutomatedBookShelfCheck(shelfURL string) {
	err := redisClient.Set(ctx, AUTOMATED_BOOK_SHELF_CHECK, shelfURL, 0).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAutomatedBookShelfCheck() string {
	shelfURL, err := redisClient.Get(ctx, AUTOMATED_BOOK_SHELF_CHECK).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return shelfURL
}

func SetDiscordWebhookURL(webhookURL string) {
	err := redisClient.Set(ctx, DISCORD_WEBHOOK_URL, webhookURL, 0).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetDiscordWebhookURL() string {
	shelfURL, err := redisClient.Get(ctx, DISCORD_WEBHOOK_URL).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return shelfURL
}

func SetDiscordMessageFormat(format string) {
	err := redisClient.Set(ctx, DISCORD_MESSAGE_FORMAT, format, 0).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetDiscordMessageFormat() string {
	shelfURL, err := redisClient.Get(ctx, DISCORD_MESSAGE_FORMAT).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return shelfURL
}
