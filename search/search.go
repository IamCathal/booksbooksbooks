package search

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger

	NON_ENGLISH_CHARACTER = regexp.MustCompile(`([^A-Za-z0-9 ,'\\\/\)\(-\.\[\]])`)
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func SearchAllRankFind(bookInfo dtos.BasicGoodReadsBook, searchResults []dtos.TheBookshopBook) dtos.EnchancedSearchResult {
	potentialAuthorMatches := []dtos.TheBookshopBook{}
	potentialTitleMatches := []dtos.TheBookshopBook{}

	for _, searchResult := range searchResults {
		authorAndTitleTheBookshop := fmt.Sprintf("%s %s", searchResult.Author, searchResult.Title)
		// theBookshopAuthor, theBookshopTitle := ExtractAuthorFromTheBookShopTitle(searchResult.Title)
		authorAndTitleGoodreads := fmt.Sprintf("%s %s", bookInfo.Author, bookInfo.Title)

		if fuzzy.MatchNormalizedFold(authorAndTitleGoodreads, authorAndTitleTheBookshop) {
			fmt.Printf("title match foundn\n")
			potentialTitleMatches = append(potentialTitleMatches, searchResult)
		}
		if fuzzy.MatchNormalizedFold(bookInfo.Author, searchResult.Author) {
			fmt.Printf("author match foundn\n")
			potentialAuthorMatches = append(potentialAuthorMatches, searchResult)
		}
	}

	searchResult := dtos.EnchancedSearchResult{
		SearchBook:    bookInfo,
		AuthorMatches: potentialAuthorMatches,
		TitleMatches:  potentialTitleMatches,
	}

	if len(potentialTitleMatches) >= 1 {
		logger.Sugar().Infof("%d potential title matches found for book: %+v matches: %+v",
			len(potentialAuthorMatches), bookInfo, potentialTitleMatches)
	}
	if len(potentialAuthorMatches) >= 1 {
		logger.Sugar().Infof("%d potential author matches found for book: %+v matches: %+v",
			len(potentialAuthorMatches), bookInfo, potentialTitleMatches)
	}

	if getOnlyEnglishBooks := db.GetOnlyEnglishBooks(); getOnlyEnglishBooks {
		searchResult = removeNonEnglishBooks(searchResult)
	}
	return searchResult
}

func removeNonEnglishBooks(searchResult dtos.EnchancedSearchResult) dtos.EnchancedSearchResult {
	filteredSearchResults := dtos.EnchancedSearchResult{
		SearchBook: searchResult.SearchBook,
	}

	for _, titleMatch := range searchResult.TitleMatches {
		isAuthorEnglish := isBookEnglish(titleMatch.Author)
		isTitleEnglish := isBookEnglish(titleMatch.Title)

		if isAuthorEnglish && isTitleEnglish {
			filteredSearchResults.TitleMatches = append(filteredSearchResults.TitleMatches, titleMatch)
		}
	}
	for _, authorMatch := range searchResult.AuthorMatches {
		isAuthorEnglish := isBookEnglish(authorMatch.Author)
		isTitleEnglish := isBookEnglish(authorMatch.Title)

		if isAuthorEnglish && isTitleEnglish {
			filteredSearchResults.AuthorMatches = append(filteredSearchResults.AuthorMatches, authorMatch)
		}
	}

	return filteredSearchResults
}

func isBookEnglish(bookDetail string) bool {
	return !NON_ENGLISH_CHARACTER.MatchString(bookDetail)
}

func removeUnnecessaryBitsFromTheBookshopTitle(fullTitleText string) (string, string) {
	author, title := ExtractAuthorFromTheBookShopTitle(fullTitleText)

	title = removeAllBetweenSubStrings(title, "(", ")")
	title = removeAllTextAfterFirstDashIfPossible(title)

	return strings.TrimSpace(author), strings.TrimSpace(title)
}

func removeAllBetweenSubStrings(sourceText, startSubstring, endSubstring string) string {
	startIndex := strings.Index(sourceText, startSubstring)
	endIndex := strings.Index(sourceText, endSubstring) + len(endSubstring)
	if startIndex == -1 || endIndex == -1 {
		return sourceText
	}
	return sourceText[:startIndex] + sourceText[endIndex:]
}

func removeAllTextAfterFirstDashIfPossible(sourceText string) string {
	splitByDash := strings.Split(sourceText, "-")
	if len(splitByDash) >= 1 {
		return strings.Join(splitByDash[:1], "-")
	}
	return sourceText
}

func ExtractAuthorFromTheBookShopTitle(fullBookTitle string) (string, string) {
	fullBookTitle = strings.TrimSpace(fullBookTitle)
	splitUpBySlash := strings.Split(fullBookTitle, "/")
	if len(splitUpBySlash) == 2 {
		return strings.TrimSpace(splitUpBySlash[0]), strings.TrimSpace(splitUpBySlash[1])
	}
	if len(splitUpBySlash) > 2 {
		return strings.TrimSpace(splitUpBySlash[0]), strings.TrimSpace(strings.Join(splitUpBySlash[1:], "-"))
	}

	splitUpByDash := strings.Split(fullBookTitle, "-")
	if len(splitUpByDash) == 2 {
		return strings.TrimSpace(splitUpByDash[0]), strings.TrimSpace(splitUpByDash[1])
	}
	if len(splitUpByDash) > 2 {
		return strings.TrimSpace(splitUpByDash[0]), strings.TrimSpace(strings.Join(splitUpByDash[1:], "-"))
	}

	return strings.TrimSpace(fullBookTitle), strings.TrimSpace(fullBookTitle)
}
