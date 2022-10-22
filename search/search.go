package search

import (
	"fmt"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func SearchAll(allBookSearchResults dtos.AllBookshopBooksSearchResults) dtos.AllBookshopBooksSearchResults {
	potentialMatches := make(dtos.AllBookshopBooksSearchResults)

	for key, searchResult := range allBookSearchResults {
		fmt.Printf("Searching for %s\n", key)
		searchResultMatches := dtos.BookShopBookSearchResult{}

		for _, possibleBook := range searchResult.SearchResultBooks {
			if fuzzy.Match(key, possibleBook.Title) {
				searchResultMatches.SearchResultBooks = append(searchResultMatches.SearchResultBooks, possibleBook)
			}
		}

		potentialMatches[key] = searchResultMatches
	}
	return potentialMatches
}

func SearchAllAuthorAndTitle(bookInfo dtos.BasicGoodReadsBook, searchResults []dtos.TheBookshopBook) dtos.EnchancedSearchResult {
	potentialAuthorMatches := []dtos.TheBookshopBook{}
	potentialTitleMatches := []dtos.TheBookshopBook{}

	for _, searchResult := range searchResults {
		if fuzzy.MatchFold(bookInfo.Title, searchResult.Title) {
			potentialTitleMatches = append(potentialTitleMatches, searchResult)
		}
		if fuzzy.MatchFold(bookInfo.Author, searchResult.Author) {
			potentialAuthorMatches = append(potentialAuthorMatches, searchResult)
		}
	}

	return dtos.EnchancedSearchResult{
		SearchBook:    bookInfo,
		AuthorMatches: potentialAuthorMatches,
		TitleMatchces: potentialTitleMatches,
	}
}
