package search

import (
	"fmt"
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

func TestRemoveallBetweenSubstrings(t *testing.T) {
	sourceText := "Mockingjay( Hunger Games Trilogy - Book 3 )"
	assert.Equal(t, "Mockingjay", removeAllBetweenSubStrings(sourceText, "(", ")"))
}
func TestRemoveallBetweenSubstringsDoesNothingWhenBothIndicesAreNotFound(t *testing.T) {
	sourceText := "Mockingjay( Hunger Games Trilogy - Book 3 )"
	assert.Equal(t, sourceText, removeAllBetweenSubStrings(sourceText, "+", "-"))
}

func TestRemoveallBetweenSubstringsDoesNothingWhenOneIndexIsNotFound(t *testing.T) {
	sourceText := "Mockingjay( Hunger Games Trilogy - Book 3 )"
	assert.Equal(t, sourceText, removeAllBetweenSubStrings(sourceText, "+", ")"))
}

func TestRemoveAllTextPastFirstDashIfPossibleRemovesAllTextBeyondOneDash(t *testing.T) {
	assert.Equal(t, "The History of Tom Jones ", removeAllTextAfterFirstDashIfPossible("The History of Tom Jones - HB - Heritage Press NY"))
}

func TestRemoveAllTextPastFirstDashIfPossibleRemovesAllTextWhenMoreThanThanTwoDashesAreFound(t *testing.T) {
	assert.Equal(t, "The History of Tom Jones ", removeAllTextAfterFirstDashIfPossible("The History of Tom Jones - HB - Heritage Press NY - SIGNED"))
}

func TestRemoveAllTextPastFirstDashIfPossibleDoesNothingWhenNoDashesAreFound(t *testing.T) {
	assert.Equal(t, "The History of Tom Jones", removeAllTextAfterFirstDashIfPossible("The History of Tom Jones"))
}

func TestExtractAuthorFromTitleSplitBySlash(t *testing.T) {
	rawTitleText := "Tolkien, J. R. R. / The Lord of the Rings"
	expectedAuthor := "Tolkien, J. R. R."
	expectedTitle := "The Lord of the Rings"

	author, title := ExtractAuthorFromTheBookShopTitle(rawTitleText)

	assert.Equal(t, expectedAuthor, author)
	assert.Equal(t, expectedTitle, title)
}

func TestExtractAuthorFromTitleSplitByHyphen(t *testing.T) {
	rawTitleText := "Tolkien, J. R. R. - The Lord of the Rings"
	expectedAuthor := "Tolkien, J. R. R."
	expectedTitle := "The Lord of the Rings"

	author, title := ExtractAuthorFromTheBookShopTitle(rawTitleText)

	assert.Equal(t, expectedAuthor, author)
	assert.Equal(t, expectedTitle, title)
}

func TestExtractAuthorFromTitleSplitByTwoHypens(t *testing.T) {
	rawTitleText := "Herbert, Frank - Le Messie de Dune ( FRENCH LANGUAGE PB ED) - En Francais"
	expectedAuthor := "Herbert, Frank"
	expectedTitle := "Le Messie de Dune ( FRENCH LANGUAGE PB ED) - En Francais"

	author, title := ExtractAuthorFromTheBookShopTitle(rawTitleText)

	assert.Equal(t, expectedAuthor, author)
	assert.Equal(t, expectedTitle, title)
}

func TestExtractAuthorFromTitleReturnsEverythingWhenCantSplit(t *testing.T) {
	rawTitleText := "Tolkien, J. R. R. [] The Lord of the Rings"
	expectedAuthor := rawTitleText
	expectedTitle := rawTitleText

	author, title := ExtractAuthorFromTheBookShopTitle(rawTitleText)

	assert.Equal(t, expectedAuthor, author)
	assert.Equal(t, expectedTitle, title)
}

func TestRemoveUnnecessaryBitsFromTheBookshopTitleRemovesParenthesisEnclosedText(t *testing.T) {
	testFullTitle := "The Waterfall ( Vintage Penguin PB 1974 - Originally 1969)"
	expectedTitle := "The Waterfall"

	actualTitle := removeUnnecessaryBitsFromTheBookshopTitle(testFullTitle)

	assert.Equal(t, expectedTitle, actualTitle)
}

func TestRemoveUnnecessaryBitsFromTheBookshopTitleRemovesPInformationBeyondFirstDash(t *testing.T) {
	testFullTitle := "Historic Murders of South Cork - SIGNED PB - BRAND NEW"
	expectedTitle := "Historic Murders of South Cork"

	actualTitle := removeUnnecessaryBitsFromTheBookshopTitle(testFullTitle)

	assert.Equal(t, expectedTitle, actualTitle)
}

func TestRemoveUnnecessaryBitsFromTheBookshopTitleRemovesPInformationBeyondFirstDashAndParenthesisEnclosedText(t *testing.T) {
	testFullTitle := "Historic Murders of South Cork - SIGNED PB - BRAND NEW - 2021 ( Murder Most Local - Book 4 ) "
	expectedTitle := "Historic Murders of South Cork"

	actualTitle := removeUnnecessaryBitsFromTheBookshopTitle(testFullTitle)

	assert.Equal(t, expectedTitle, actualTitle)
}

func TestRemoveEmptyStringElementsInArr(t *testing.T) {
	expectedArr := []string{"hello", "world"}
	assert.DeepEqual(t, expectedArr, removeEmptyStringElementsInArr([]string{"hello", "", "world"}))
}

func TestRemoveEmptyStringElementsInArrDoesNothingWhenNoEmptyElementsAreGiven(t *testing.T) {
	expectedArr := []string{"hello", "world"}
	assert.DeepEqual(t, expectedArr, removeEmptyStringElementsInArr([]string{"hello", "world"}))
}

func TestGetAuthorFullNameInCorrectOrderForPeterFHamilton(t *testing.T) {
	assert.Equal(t, "Peter F. Hamilton", getAuthorFullNameInCorrectOrder("Hamilton, Peter F."))
}

func TestGetAuthorFullNameInCorrectOrderForPatrickRothfuss(t *testing.T) {
	assert.Equal(t, "Patrick Rothfuss", getAuthorFullNameInCorrectOrder("Rothfuss, Patrick"))
}

func TestGetAuthorFullNameInCorrectOrderForJohnWCampbellJr(t *testing.T) {
	assert.Equal(t, "John W. Campbell Jr.", getAuthorFullNameInCorrectOrder("Campbell Jr., John W. "))
}

func TestTokeniseTitle(t *testing.T) {
	expectedTokenisedTitle := []string{"the", "name", "of", "the", "wind"}
	assert.DeepEqual(t, expectedTokenisedTitle, tokeniseTitle("the name of the wind"))
}

func TestTokeniseTitleHandlesMultipleSpaces(t *testing.T) {
	expectedTokenisedTitle := []string{"the", "name", "of", "the", "wind"}
	assert.DeepEqual(t, expectedTokenisedTitle, tokeniseTitle("the name of               the wind"))
}

func TestTokeniseTitleLowercasesAllTokens(t *testing.T) {
	expectedTokenisedTitle := []string{"the", "name", "of", "the", "wind"}
	assert.DeepEqual(t, expectedTokenisedTitle, tokeniseTitle("THE NamE of THe WINd"))
}

func TestLowercaseAllStringElement(t *testing.T) {
	assert.DeepEqual(t, []string{"hello"}, lowercaseAllStringElements([]string{"heLLO"}))
}

func TestTitleMatch(t *testing.T) {
	searchBookTokens := []string{"the", "name", "of", "the", "wind"}
	searchResultTokens := []string{"the", "name", "of", "the", "wind"}

	assert.Equal(t, true, titlesMatch(searchBookTokens, searchResultTokens))
}

func TestGetTheBookshopTitleGetsRidOfTextInParethesis(t *testing.T) {
	assert.Equal(t, "The Crystal Run", getPureTheBookshopTitle("The Crystal Run (Signed by the Author) "))
}

func TestGetTheBookshopTitleGetsRidOfTextAfterTooManyDashes(t *testing.T) {
	assert.Equal(t, "Black Juice", getPureTheBookshopTitle("Black Juice - HB - Gollancz - Short Stories"))
}

func TestGetTheBookshopTitleDoesNothingToARegularTitle(t *testing.T) {
	assert.Equal(t, "Teranesia", getPureTheBookshopTitle("Teranesia"))
}

func TestGetTheBookshopTitleRemovesParentesisTextFirst(t *testing.T) {
	assert.Equal(t, "Mockingjay", getPureTheBookshopTitle("Mockingjay ( Hunger Games Trilogy - Book 3 )"))
}

func TestBookWithLongTitleDoesntMatchWithSmallerSubstringMatch(t *testing.T) {
	resetDBFields()

	searchBook := dtos.BasicGoodReadsBook{
		Title:  "How To Fall In Love",
		Author: "Ahern, Cecelia",
	}
	searchResultsFromTheBookShop := []dtos.TheBookshopBook{
		{
			Title:  "Fall In Love",
			Author: "Ahern, Cecelia",
		},
		{
			Title:  fmt.Sprintf("%s.", searchBook.Title),
			Author: "Ahern, Cecelia",
		},
	}

	searchResult := SearchAllRankFind(searchBook, searchResultsFromTheBookShop)

	assert.Equal(t, 0, len(searchResult.TitleMatches))
	assert.Equal(t, len(searchResultsFromTheBookShop), len(searchResult.AuthorMatches))
}

func resetDBFields() {
	db.SetKnownAuthors([]dtos.KnownAuthor{})
	db.SetAddMoreAuthorBooksToAvailableBooksList(false)
	db.SetSendAlertWhenBookNoLongerAvailable(false)
	db.SetOnlyEnglishBooks(false)
	db.SetAvailableBooks([]dtos.AvailableBook{})
	db.SetDiscordWebhookURL("")
}
