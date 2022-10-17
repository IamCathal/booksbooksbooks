package goodreads

import (
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
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
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
