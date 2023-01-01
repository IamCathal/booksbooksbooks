package search

import (
	"log"
	"os"
	"testing"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"go.uber.org/zap"
	"gotest.tools/assert"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	SetLogger(logger)
	db.SetLogger(logger)

	db.ConnectToRedis()
	db.SetTestDataIdentifiers()

	code := m.Run()

	os.Exit(code)
}

func TestSearchCorrectlyExtractsActualAuthorMatches(t *testing.T) {
	resetDBFields()
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

	assert.DeepEqual(t, expectedAuthorMatches, actualSearchResult.AuthorMatches)
}

func TestSearchCorrectlyExtractsActualTitleMatches(t *testing.T) {
	resetDBFields()
	wiseMansFearResult := dtos.TheBookshopBook{
		Author: "paTRICK ROTHFUSS",
		Title:  "THE WISE MANS FEAR",
	}
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
		wiseMansFearResult,
	}

	actualSearchResult := SearchAllRankFind(searchBookInfo, searchResults)

	assert.DeepEqual(t, []dtos.TheBookshopBook{wiseMansFearResult}, actualSearchResult.TitleMatches)
}

func TestSearchParametersIgnoreNonAlphaNumericSymbols(t *testing.T) {
	resetDBFields()
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

	assert.DeepEqual(t, searchResults, actualSearchResult.TitleMatches)
}

func TestSearchAllRankFindDoesNotReturnNonEnglishBooksWhenSettingIsEnabled(t *testing.T) {
	resetDBFields()
	searchBook := dtos.BasicGoodReadsBook{
		Title:  "The Wise Man's Fear",
		Author: "Patrick Rothfuss",
	}
	searchResults := []dtos.TheBookshopBook{
		{
			Title:  "The Wise Man's Fear",
			Author: "Patrick Rothfuss",
		},
		{
			Title:  "имя ветра",
			Author: "Patrick Rothfuss",
		},
	}
	db.SetOnlyEnglishBooks(true)

	searchResult := SearchAllRankFind(searchBook, searchResults)

	assert.Equal(t, 1, len(searchResult.TitleMatches))
	assert.Equal(t, 1, len(searchResult.AuthorMatches))
}

func TestSearchAllRankFindDoesReturnNonEnglishBooksWhenSettingIsDisabled(t *testing.T) {
	resetDBFields()
	searchBook := dtos.BasicGoodReadsBook{
		Title:  "The Wise Man's Fear",
		Author: "Patrick Rothfuss",
	}
	searchResults := []dtos.TheBookshopBook{
		{
			Title:  "имя ветра",
			Author: "Patrick Rothfuss",
		},
	}
	db.SetOnlyEnglishBooks(false)

	searchResult := SearchAllRankFind(searchBook, searchResults)

	assert.DeepEqual(t, []dtos.TheBookshopBook{}, searchResult.TitleMatches)
	assert.DeepEqual(t, searchResults, searchResult.AuthorMatches)
}

func TestIsBookEnglishDetectsBookWithFrenchFada(t *testing.T) {
	assert.Equal(t, false, isBookEnglish("Parrot, André - Sumer - FRENCH LANGUAGE Edition"))
}

func TestIsBookEnglishDetectsBookWithAUmlaut(t *testing.T) {
	assert.Equal(t, false, isBookEnglish("Doerr, Anthony -Kaikki se valo jota emme näe - HB - Finnish"))
}
func TestIsBookEnglishDetectsBookWithPolishFancyZ(t *testing.T) {
	assert.Equal(t, false, isBookEnglish("McCaffrey, Anne -Historia Nerilki ( Jeźdźcy smoków z"))
}
func TestIsBookEnglishDetectsBookWithCyrillicLetters(t *testing.T) {
	assert.Equal(t, false, isBookEnglish("имя ветра"))
}
func TestIsBookEnglishDetectsBookWithFSharphesS(t *testing.T) {
	assert.Equal(t, false, isBookEnglish("Schiller, Friedrich - Geschichte des dreißigjährigen Kriegs"))
}
func TestIsBookEnglishDetectsEnglishTitleBook(t *testing.T) {
	assert.Equal(t, true, isBookEnglish("Collins, Suzanne / The Hunger Games ( Hunger Games Trilogy "))
}

func resetDBFields() {
	db.SetKnownAuthors([]dtos.KnownAuthor{})
	db.SetAddMoreAuthorBooksToAvailableBooksList(false)
	db.SetSendAlertWhenBookNoLongerAvailable(false)
	db.SetOnlyEnglishBooks(false)
	db.SetAvailableBooks([]dtos.AvailableBook{})
	db.SetDiscordWebhookURL("")
}
