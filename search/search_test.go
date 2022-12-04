package search

import (
	"log"
	"os"
	"testing"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)

	code := m.Run()

	os.Exit(code)
}

func TestSearchCorrectlyExtractsActualAuthorMatches(t *testing.T) {
	nameOfTheWind := dtos.TheBookshopBook{
		Author: "Patrick rothfuss",
		Title:  "the name of the wind",
	}
	wiseMansFear := dtos.TheBookshopBook{
		Author: "paTRICK ROTHFUSS",
		Title:  "the wise mans fear",
	}

	searchBookInfo := dtos.BasicGoodReadsBook{
		Author: "Patrick Rothfuss",
		Title:  "the wise man's fear",
	}
	searchResults := []dtos.TheBookshopBook{
		{
			Author: "Swayze, Patrick",
			Title:  "The Time of My life",
		},
		{
			Author: "St Aubyn, Edward",
			Title:  "Patrick Melrose volume 1",
		},
		nameOfTheWind,
		wiseMansFear,
	}
	expectedAuthorMatches := []dtos.TheBookshopBook{
		nameOfTheWind, wiseMansFear,
	}

	actualSearchResult := SearchAllRankFind(searchBookInfo, searchResults)

	assert.Equal(t, expectedAuthorMatches, actualSearchResult.AuthorMatches)
}

func TestSearchCorrectlyExtractsActualTitleMatches(t *testing.T) {
	searchBookInfo := dtos.BasicGoodReadsBook{
		Author: "Patrick Rothfuss",
		Title:  "the wise mans fear",
	}
	searchResults := []dtos.TheBookshopBook{
		{
			Author: "Swayze, Patrick",
			Title:  "The Time of My life",
		},
		{
			Author: "St Aubyn, Edward",
			Title:  "Patrick Melrose volume 1",
		},
		{
			Author: "paTRICK ROTHFUSS",
			Title:  "THE WISE MANS FEAR",
		},
	}

	actualSearchResult := SearchAllRankFind(searchBookInfo, searchResults)

	assert.Len(t, actualSearchResult.TitleMatches, 1)
}

func TestSearchParametersIgnoreNonAlphaNumericSymbols(t *testing.T) {
	searchBookInfo := dtos.BasicGoodReadsBook{
		Author: "Patrick Rothfuss",
		Title:  "the wise mans fear",
	}
	searchResults := []dtos.TheBookshopBook{
		{
			Author: "paTRICK ROTHFUSS",
			Title:  "THE '''''''''''''''''WISE MAN''''''''''''''''''''''S FEAR",
		},
	}

	actualSearchResult := SearchAllRankFind(searchBookInfo, searchResults)

	assert.Len(t, actualSearchResult.TitleMatches, 1)
}
