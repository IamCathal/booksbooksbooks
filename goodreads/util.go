package goodreads

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	// There are five spaces between a books
	// title and its series information if
	// the series information is given
	TITLE_AND_SERIES_INFO_SEPERATOR = regexp.MustCompile("[ ]{3,}")
	// Goodreads returns 30 books per page
	BOOK_COUNT_PER_PAGE = 30
	// Crude to check if a roughly  valid
	// shelf URL is being queried
	GOODREADS_SHELF_URL_PREFIX = "https://www.goodreads.com/review/list/"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func checkIsShelfURL(checkURL string) bool {
	hasPrefix := strings.HasPrefix(checkURL, GOODREADS_SHELF_URL_PREFIX)
	properURL, err := url.Parse(checkURL)
	checkErr(err)
	shelfParam := properURL.Query().Get("shelf")

	return hasPrefix && shelfParam != ""
}

func processBook(fullTitle, author string) dtos.BasicGoodReadsBook {
	fullTitle = stripOfFormatting(fullTitle)
	author = stripOfFormatting(author)
	bookTitle, seriesInfo := extractTitleDetailsIfPossible(fullTitle)
	newBook := dtos.BasicGoodReadsBook{
		Title:      bookTitle,
		Author:     author,
		SeriesText: seriesInfo,
	}
	return newBook
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
	panic(splitBySpace)
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
