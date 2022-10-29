package search

import (
	"fmt"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func SearchAllRankFind(bookInfo dtos.BasicGoodReadsBook, searchResults []dtos.TheBookshopBook) dtos.EnchancedSearchResult {
	potentialAuthorMatches := []dtos.TheBookshopBook{}
	potentialTitleMatches := []dtos.TheBookshopBook{}

	for _, searchResult := range searchResults {
		titleAndAuthorTheBookshop := fmt.Sprintf("%s %s", searchResult.Author, searchResult.Title)
		titleAndAuthorGoodReads := fmt.Sprintf("%s %s", bookInfo.Author, bookInfo.Title)

		if fuzzy.MatchFold(titleAndAuthorGoodReads, titleAndAuthorTheBookshop) {
			potentialTitleMatches = append(potentialTitleMatches, searchResult)
		}
		if fuzzy.MatchFold(bookInfo.Author, searchResult.Author) {
			potentialAuthorMatches = append(potentialAuthorMatches, searchResult)
		}
	}

	if len(potentialTitleMatches) >= 2 {
		logger.Sugar().Infof("%d potential title matches found for book: %+v matches: %+v",
			len(potentialAuthorMatches), bookInfo, potentialTitleMatches)
	}
	return dtos.EnchancedSearchResult{
		SearchBook:    bookInfo,
		AuthorMatches: potentialAuthorMatches,
		TitleMatches:  potentialTitleMatches,
	}
}
