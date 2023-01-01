package search

import (
	"fmt"
	"regexp"

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
		titleAndAuthorTheBookshop := fmt.Sprintf("%s %s", searchResult.Author, searchResult.Title)
		titleAndAuthorGoodreads := fmt.Sprintf("%s %s", bookInfo.Author, bookInfo.Title)

		if fuzzy.MatchNormalizedFold(titleAndAuthorGoodreads, titleAndAuthorTheBookshop) {
			potentialTitleMatches = append(potentialTitleMatches, searchResult)
		}
		if fuzzy.MatchNormalizedFold(bookInfo.Author, searchResult.Author) {
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
