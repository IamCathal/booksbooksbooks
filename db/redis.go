package db

import (
	"context"
	"encoding/json"
	"fmt"

	redis "github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	ctx         = context.Background()
	redisClient *redis.Client

	AVAILABLE_BOOKS = "availableBooks"
	RECENT_CRAWLS   = "recentCrawls"
)

func ConnectToRedis() {
	fmt.Printf("Connecting to redis...\n")

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	redisClient = rdb
	response, err := redisClient.Ping(ctx).Result()
	if err != nil {
		panic(response)
	}

	fmt.Printf("Redis connection successfully initialised\n")
}

func ResetAvailableBooks() {
	err := redisClient.Set(ctx, AVAILABLE_BOOKS, []byte(""), 0).Err()
	if err != nil {
		panic(err)
	}
}

func AddAvailableBook(newBook dtos.AvailableBook) {
	availableBooks := GetAvailableBooks()
	// fmt.Printf("curr availableBooks %d %+v\n\n", len(availableBooks), availableBooks)
	availableBooks = append(availableBooks, newBook)
	jsonAvailableBooks, err := json.Marshal(availableBooks)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("Adding new available book: %+v\n", newBook)
	err = redisClient.Set(ctx, AVAILABLE_BOOKS, jsonAvailableBooks, 0).Err()
	if err != nil {
		panic(err)
	}
}

func GetAvailableBooks() []dtos.AvailableBook {
	availableBooksStr, err := redisClient.Get(ctx, AVAILABLE_BOOKS).Result()
	if err == redis.Nil {
		return []dtos.AvailableBook{}
	} else if err != nil {
		panic(err)
	}
	availableBooks := []dtos.AvailableBook{}
	if availableBooksStr != "" {
		err = json.Unmarshal([]byte(availableBooksStr), &availableBooks)
		if err != nil {
			panic(err)
		}
	}
	return availableBooks
}

func GetRecentCrawls() []dtos.RecentCrawl {
	recentCrawls, err := redisClient.Get(ctx, RECENT_CRAWLS).Result()
	if err == redis.Nil {
		return []dtos.RecentCrawl{}
	} else if err != nil {
		panic(err)
	}
	recentCrawlsArr := []dtos.RecentCrawl{}
	if recentCrawls != "" {
		err = json.Unmarshal([]byte(recentCrawls), &recentCrawlsArr)
		if err != nil {
			panic(err)
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
		panic(err)
	}
	err = redisClient.Set(ctx, RECENT_CRAWLS, jsonCrawls, 0).Err()
	if err != nil {
		panic(err)
	}
}
