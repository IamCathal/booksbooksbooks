package search

import (
	"fmt"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func SearchAll(allBookSearchResults dtos.AllBookshopBooksSearchResults) []dtos.TheBookshopBook {
	potentialMatches := []dtos.TheBookshopBook{}

	for key, searchResult := range allBookSearchResults {
		fmt.Printf("Searching for %s\n", key)

		for _, possibleBook := range searchResult.SearchResultBooks {
			if fuzzy.Match(key, possibleBook.Title) {
				potentialMatches = append(potentialMatches, possibleBook)
			}
		}
	}
	return potentialMatches
}
