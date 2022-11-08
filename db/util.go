package db

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func GetAvailableBooksMap() map[string]bool {
	availableBooks := GetAvailableBooks()
	availableBooksMap := make(map[string]bool)

	for _, book := range availableBooks {
		availableBooksMap[book.BookPurchaseInfo.Link] = true
	}
	return availableBooksMap
}

func getKeyForRecentCrawlBreadcrumb(shelfURL string) string {
	urlObj, err := url.Parse(shelfURL)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	name := strings.Split(urlObj.Path, "-")
	return fmt.Sprintf("%s-%s", name[len(name)-1], urlObj.Query().Get("shelf"))
}

func IgnoreBook(bookURL string) {
	newAvailableBooks := []dtos.AvailableBook{}
	for _, book := range GetAvailableBooks() {
		if book.BookPurchaseInfo.Link == bookURL {
			book.Ignore = true
		}
		newAvailableBooks = append(newAvailableBooks, book)
	}
	SetAvailableBooks(newAvailableBooks)
}

func UnignoreBook(bookURL string) {
	newAvailableBooks := []dtos.AvailableBook{}
	for _, book := range GetAvailableBooks() {
		if book.BookPurchaseInfo.Link == bookURL {
			book.Ignore = false
		}
		newAvailableBooks = append(newAvailableBooks, book)
	}
	SetAvailableBooks(newAvailableBooks)
}

func removeDuplicateAvailableBooks(books []dtos.AvailableBook) []dtos.AvailableBook {
	seenBooks := make(map[string]bool)
	noDuplicateAvailableBooks := []dtos.AvailableBook{}

	for _, book := range books {
		_, exists := seenBooks[book.BookPurchaseInfo.Link]
		if !exists {
			seenBooks[book.BookInfo.Title] = true
			noDuplicateAvailableBooks = append(noDuplicateAvailableBooks, book)
		}
	}
	return noDuplicateAvailableBooks
}

func removeDuplicateRecentCrawls(recentCrawls []dtos.RecentCrawlBreadcrumb) []dtos.RecentCrawlBreadcrumb {
	seenBreadcrumbs := make(map[string]bool)
	noDuplicateRecentCrawlBreadcrumbs := []dtos.RecentCrawlBreadcrumb{}

	for _, crawl := range recentCrawls {
		_, exists := seenBreadcrumbs[crawl.ShelfURL]
		if !exists {
			seenBreadcrumbs[crawl.ShelfURL] = true
			noDuplicateRecentCrawlBreadcrumbs = append(noDuplicateRecentCrawlBreadcrumbs, crawl)
		}
	}
	return noDuplicateRecentCrawlBreadcrumbs
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
		logger.Sugar().Fatal(err)
	}
	return canBuy
}

func isRedisNil(err error) bool {
	return err == redis.Nil
}

func strToBool(stringBool string) bool {
	boolVal, err := strconv.ParseBool(stringBool)
	if err != nil {
		// logger.Sugar().Fatalf("failed to parse '%s' to bool", stringBool)
		panic(err)
	}
	return boolVal
}
