package db

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/segmentio/ksuid"
)

var (
	ANYTHING_BUT_LETTERS_AND_SPACES = regexp.MustCompile(`[^A-Za-z0-9 ]`)
)

func GetAvailableBooksMap() map[string]bool {
	availableBooks := GetAvailableBooks()
	availableBooksMap := make(map[string]bool)

	for _, book := range availableBooks {
		availableBooksMap[book.BookPurchaseInfo.Link] = true
	}
	return availableBooksMap
}

func GetMapForAvailableBooks(availableBooks []dtos.AvailableBook) map[string]bool {
	availableBooksMap := make(map[string]bool)

	for _, book := range availableBooks {
		availableBooksMap[book.BookPurchaseInfo.Link] = true
	}
	return availableBooksMap
}

func GetKeyForRecentCrawlBreadcrumb(shelfURL string) string {
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

func removeDuplicateRecentCrawls(recentCrawlBreadcrumbs []dtos.RecentCrawlBreadcrumb) []dtos.RecentCrawlBreadcrumb {
	seenBreadcrumbs := make(map[string]bool)
	noDuplicateRecentCrawlBreadcrumbs := []dtos.RecentCrawlBreadcrumb{}

	for _, crawl := range recentCrawlBreadcrumbs {
		_, exists := seenBreadcrumbs[crawl.ShelfURL]
		if !exists {
			seenBreadcrumbs[crawl.ShelfURL] = true
			noDuplicateRecentCrawlBreadcrumbs = append(noDuplicateRecentCrawlBreadcrumbs, crawl)
		}
	}
	return noDuplicateRecentCrawlBreadcrumbs
}

func removeDuplicateAuthors(authors []dtos.KnownAuthor) []dtos.KnownAuthor {
	seenAuthors := make(map[string]bool)
	noDuplicateRecentAuthors := []dtos.KnownAuthor{}

	for _, author := range authors {
		_, exists := seenAuthors[author.Name]
		if !exists {
			seenAuthors[author.Name] = true
			noDuplicateRecentAuthors = append(noDuplicateRecentAuthors, author)
		}
	}
	return noDuplicateRecentAuthors
}

func removeDuplicateAuthorsDisregardingReverseOrder(authors []dtos.KnownAuthor) []dtos.KnownAuthor {
	// "Collins, Suzanne" isn't an exact duplicate of
	// of "Suzanne Collins" but to a human it is. This is quite annoying
	// in terms of writing up good enough solution
	noDuplicateAuthors := []dtos.KnownAuthor{}
	seenAuthorSortedStringNames := make(map[string]bool)

	for _, author := range authors {
		sortedStringName := convertAuthorNameToSortedString(author.Name)
		if _, seen := seenAuthorSortedStringNames[sortedStringName]; !seen {
			seenAuthorSortedStringNames[sortedStringName] = true
			noDuplicateAuthors = append(noDuplicateAuthors, author)
		}
	}
	return noDuplicateAuthors
}

func convertAuthorNameToSortedString(authorNameRaw string) string {
	authorNameTokens := getAuthorNameTokens(authorNameRaw)
	sort.Strings(authorNameTokens)
	return strings.Join(authorNameTokens, "")
}

func getAuthorNameTokens(authorNameRaw string) []string {
	onlyLettersAndSpaces := string(ANYTHING_BUT_LETTERS_AND_SPACES.ReplaceAll([]byte(authorNameRaw), []byte("")))
	tokens := strings.Split(onlyLettersAndSpaces, " ")
	return removeEmptyStringElementsInArr(tokens)
}

func removeEmptyStringElementsInArr(arr []string) []string {
	noEmptyElementsArr := []string{}
	for _, elem := range arr {
		if elem != "" {
			noEmptyElementsArr = append(noEmptyElementsArr, elem)
		}
	}
	return noEmptyElementsArr
}

func GetIgnoredAndNonIgnoredCountOfAvailableBooks() (int, int) {
	allAvailableBooks := GetAvailableBooks()
	nonIgnoredCount := 0
	ignoredCount := 0

	for _, book := range allAvailableBooks {
		if book.Ignore {
			ignoredCount++
		} else {
			nonIgnoredCount++
		}
	}
	return nonIgnoredCount, ignoredCount
}

func strToBool(stringBool string) bool {
	boolVal, err := strconv.ParseBool(stringBool)
	if err != nil {
		panic(err)
	}
	return boolVal
}

func strToInt(stringInt string) int {
	intVal, err := strconv.Atoi(stringInt)
	if err != nil {
		panic(err)
	}
	return intVal
}

func IsIgnoredAuthor(author string) bool {
	ignoredAuthors := getIgnoredAuthors()
	for _, ignoredAuthor := range ignoredAuthors {
		if ignoredAuthor == author {
			return true
		}
	}
	return false
}

func getShelvesWithoutDuplicates(shelves []dtos.ShelfToCrawl) []dtos.ShelfToCrawl {
	shelvesWithoutDupicates := []dtos.ShelfToCrawl{}
	seenShelves := make(map[string]bool)

	for _, shelf := range shelves {
		if _, seen := seenShelves[shelf.ShelfURL]; !seen {
			seenShelves[shelf.ShelfURL] = true
			shelvesWithoutDupicates = append(shelvesWithoutDupicates, shelf)
		}
	}
	return shelvesWithoutDupicates
}

func GetShelfURLsFromShelvesToCrawl() []string {
	URLs := []string{}
	for _, shelfToCrawl := range GetShelvesToCrawl() {
		URLs = append(URLs, shelfToCrawl.ShelfURL)
	}
	return URLs
}

func GetShelfCrawlKeysFromShelvesToCrawl() []string {
	URLs := []string{}
	for _, shelfToCrawl := range GetShelvesToCrawl() {
		URLs = append(URLs, shelfToCrawl.CrawlKey)
	}
	return URLs
}

func SetTestDataIdentifiers() {
	randomID := ksuid.New().String()
	AVAILABLE_BOOKS = "test-" + randomID + "-availableBooks"
	RECENT_CRAWL_BREADCRUMBS = "test-" + randomID + "-recentCrawls"
	AUTOMATED_BOOK_SHELF_CRAWL_TIME = "test-" + randomID + "-automatedBookShelfCrawlTime"
	DISCORD_WEBHOOK_URL = "test-" + randomID + "-discordWebHookURL"
	DISCORD_MESSAGE_FORMAT = "test-" + randomID + "-discordMessageFormat"
	SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE = "test-" + randomID + "-sendAlertWhenBookNoLongerAvailable"
	SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN = "test-" + randomID + "-sendAlertWhenFreeShippingKicksIn"
	TOTAL_BOOKS_IN_AUTOMATED_BOOK_SHELF = "test-" + randomID + "-totalBooksInAutomatedBookShelf"
	ADD_MORE_AUTHOR_BOOKS_TO_AVAILABLE_BOOKS_LIST = "test-" + randomID + "-addMoreAuthorBooksToAvailableBooksList"
	KNOWN_AUTHORS = "test-" + randomID + "-knownAuthors"
	IGNORE_AUTHORS = "test-" + randomID + "-ignoreAuthors"
	OTHER_BOOKS_IN_SERIES_LOOKUP = "test-" + randomID + "-otherBooksInSeriesLookup"
	OWNED_BOOKS_SHELF_URL = "test-" + randomID + "-ownedBooksShelfURL"
	SERIES_CRAWL_BOOKS = "test-" + randomID + "-seriesCrawlBooks"
	ONLY_ENGLISH_BOOKS_TOGGLE = "test-" + randomID + "-onlyEnglishBooksToggle"
	SHELVES_TO_CRAWL = "test-" + randomID + "-shelvesToCrawl"
	DEFAULT_TTL = time.Duration(300 * time.Second)
}
