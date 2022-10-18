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
