package db

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func GetAvailableBooksMap() map[string]bool {
	availableBooks := GetAvailableBooks()
	availableBooksMap := make(map[string]bool)

	for _, book := range availableBooks {
		availableBooksMap[book.BookPurchaseInfo.Title] = true
	}
	return availableBooksMap
}

func getKeyForRecentCrawl(shelfURL string) string {
	urlObj, err := url.Parse(shelfURL)
	if err != nil {
		panic(err)
	}
	name := strings.Split(urlObj.Path, "-")
	return fmt.Sprintf("%s-%s", name[len(name)-1], urlObj.Query().Get("shelf"))
}

func removeDuplicateRecentCrawls(recentCrawls []dtos.RecentCrawl) []dtos.RecentCrawl {
	seenShelves := make(map[string]bool)
	noDuplicateRecentCrawls := []dtos.RecentCrawl{}

	for _, crawl := range recentCrawls {
		_, exists := seenShelves[crawl.ShelfURL]
		if !exists {
			seenShelves[crawl.ShelfURL] = true
			noDuplicateRecentCrawls = append(noDuplicateRecentCrawls, crawl)
		}
	}
	return noDuplicateRecentCrawls
}

func getAppropriateID(book dtos.BasicGoodReadsBook) string {
	if book.Isbn13 != "" {
		return book.Isbn13
	} else if book.Asin != "" {
		return book.Asin
	}
	return fmt.Sprintf("%s/%s", book.Author, book.Title)
}

func getCurrentBookState(book dtos.BasicGoodReadsBook) string {
	id := getAppropriateID(book)
	canBuy, err := redisClient.Get(ctx, id).Result()
	if err != nil && isRedisNil(err) {
		panic(err)
	}
	return canBuy
}

func isRedisNil(err error) bool {
	return err == redis.Nil
}
