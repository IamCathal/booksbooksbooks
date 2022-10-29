package goodreads

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/segmentio/ksuid"
)

var (
	// There are five spaces between a books
	// title and its series information if
	// the series information is given
	TITLE_AND_SERIES_INFO_SEPERATOR = regexp.MustCompile("[ ]{3,}")
	// Goodreads returns 30 books per page
	BOOK_COUNT_PER_PAGE = 30
	// Base URL that book links are built on
	GOODREADS_BASE_BOOK_URL = "https://www.goodreads.com"
	// Crude to check if a roughly  valid
	// shelf URL is being queried
	GOODREADS_SHELF_URL_PREFIX = GOODREADS_BASE_BOOK_URL + "/review/list/"
)

func checkErr(err error) {
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func CheckIsShelfURL(checkURL string) bool {
	hasPrefix := strings.HasPrefix(checkURL, GOODREADS_SHELF_URL_PREFIX)
	properURL, err := url.Parse(checkURL)
	checkErr(err)
	shelfParam := properURL.Query().Get("shelf")

	return hasPrefix && shelfParam != ""
}

func processBook(fullTitle, author, cover, isbn13, asin, rating, link string) dtos.BasicGoodReadsBook {
	fullTitle = stripOfFormatting(fullTitle)
	author = stripOfFormatting(author)
	cover = stripOfFormatting(cover)
	isbn13 = stripOfFormatting(isbn13)
	asin = stripOfFormatting(asin)
	rating = stripOfFormatting(rating)
	link = GOODREADS_BASE_BOOK_URL + link

	value, err := strconv.ParseFloat(rating, 32)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	bookTitle, seriesInfo := extractTitleDetailsIfPossible(fullTitle)
	newBook := dtos.BasicGoodReadsBook{
		ID:         ksuid.New().String(),
		Title:      bookTitle,
		Author:     author,
		SeriesText: seriesInfo,
		Link:       link,
		Cover:      cover,
		Isbn13:     isbn13,
		Asin:       asin,
		Rating:     float32(value),
	}
	return newBook
}

func GetAvailableBooksFromSearchResult(searchResults []dtos.EnchancedSearchResult) []dtos.AvailableBook {
	availableBooks := []dtos.AvailableBook{}
	for _, searchResult := range searchResults {
		if len(searchResult.TitleMatches) >= 1 {
			availableBook := dtos.AvailableBook{
				BookInfo:         searchResult.SearchBook,
				BookPurchaseInfo: searchResult.TitleMatches[0],
			}
			availableBooks = append(availableBooks, availableBook)
		}
	}
	return availableBooks
}

func stripOfFormatting(input string) string {
	formatted := strings.ReplaceAll(input, "\n", "")
	formatted = strings.TrimSpace(formatted)
	return formatted
}

func extractTitleDetailsIfPossible(fullTitle string) (string, string) {
	splitFullTitle := TITLE_AND_SERIES_INFO_SEPERATOR.Split(fullTitle, 2)
	if len(splitFullTitle) == 2 {
		return splitFullTitle[0], splitFullTitle[1]
	}
	return fullTitle, ""
}

func extractLoadedCount(loadedCountText string) (int, int) {
	loadedCountText = strings.TrimSpace(loadedCountText)
	splitBySpace := strings.Split(loadedCountText, " ")
	if len(splitBySpace) == 4 {
		return strToInt(splitBySpace[0]), strToInt(splitBySpace[2])
	}
	logger.Sugar().Fatal(splitBySpace)
	return 0, 0
}

func strToInt(str string) int {
	intVersion, err := strconv.Atoi(str)
	checkErr(err)
	return intVersion
}

func totalPagesToCrawl(totalBooks int) int {
	fullPages, nonFullPageIfMoreThanOne := divmod(totalBooks, BOOK_COUNT_PER_PAGE)
	if (nonFullPageIfMoreThanOne) >= 1 {
		return fullPages + 1
	}
	return fullPages
}

func divmod(big, little int) (int, int) {
	quotient := big / little
	remainder := big % little
	return quotient, remainder
}

func getFakeReferrerPage(URL string) string {
	splitStringByShelfParam := strings.Split(URL, "?")
	return splitStringByShelfParam[0]
}
