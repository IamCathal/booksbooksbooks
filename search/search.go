package search

import (
	"fmt"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func Search(goodReadsBook dtos.BasicGoodReadsBook, theBookshopResults []dtos.TheBookshopBook) []dtos.TheBookshopBook {
	potentialMatches := []dtos.TheBookshopBook{}
	fmt.Printf("Searching for %s\n", goodReadsBook.Title)

	for _, searchResult := range theBookshopResults {
		if fuzzy.Match(goodReadsBook.Title, searchResult.Title) {
			potentialMatches = append(potentialMatches, searchResult)
		}
	}

	return potentialMatches
}
