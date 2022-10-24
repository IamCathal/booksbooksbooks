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
)

func ConnectToRedis() {
	fmt.Printf("Connecting to redis...\n")

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	redisClient = rdb

	// err := redisClient.Set(ctx, "key", "value", 0).Err()
	// if err != nil {
	// 	panic(err)
	// }

	// val, err := redisClient.Get(ctx, "key").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("key", val)

	// val2, err := rdb.Get(ctx, "key2").Result()
	// if err == redis.Nil {
	// 	fmt.Println("key2 does not exist")
	// } else if err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("key2", val2)
	// }
	fmt.Printf("Redis connection successfully initialised\n")
}

func GetRecentCrawls() []dtos.RecentCrawl {
	recentCrawls, err := redisClient.Get(ctx, "recentCrawls").Result()
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
	recentCrawls, err := redisClient.Get(ctx, "recentCrawls").Result()
	if err != nil && !isNotRedisNil(err) {
		panic(err)
	}

	recentCrawlsArr := []dtos.RecentCrawl{}
	if recentCrawls != "" {
		err = json.Unmarshal([]byte(recentCrawls), &recentCrawlsArr)
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("The recent crawls: %+v\n", recentCrawlsArr)

	setNewRecentCrawls := []dtos.RecentCrawl{
		{
			CrawlKey: getKeyForRecentCrawl(shelfURL),
			ShelfURL: shelfURL,
		},
	}
	setNewRecentCrawls = append(setNewRecentCrawls, recentCrawlsArr...)
	setNewRecentCrawls = removeDuplicateRecentCrawls(setNewRecentCrawls)

	jsonCrawls, err := json.Marshal(setNewRecentCrawls)
	if err != nil {
		panic(err)
	}

	err = redisClient.Set(ctx, "recentCrawls", jsonCrawls, 0).Err()
	if err == redis.Nil {
		fmt.Printf("couldnt set recentCrawls\n")
	} else if err != nil {
		panic(err)
	}
}
