package db

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

	splitUrlbyDash := strings.Split(urlObj.Path, "-")
	nameComponent := ""

	if len(splitUrlbyDash) == 1 {
		splitBySlash := strings.Split(urlObj.Path, "/")
		nameComponent = splitBySlash[len(splitBySlash)-1]
	} else {
		nameComponent = splitUrlbyDash[len(splitUrlbyDash)-1]
	}

	return fmt.Sprintf("%s-%s", nameComponent, urlObj.Query().Get("shelf"))
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
			seenBooks[book.BookPurchaseInfo.Link] = true
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

func strToBool(stringBool string) bool {
	boolVal, err := strconv.ParseBool(stringBool)
	if err != nil {
		panic(err)
	}
	return boolVal
}
