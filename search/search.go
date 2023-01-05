package search

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
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

		if strings.EqualFold(bookInfo.Author, searchResult.Author) {
			potentialAuthorMatches = append(potentialAuthorMatches, searchResult)

			if strings.Contains(searchResult.Title, ":") {
				continue
			}

			theBookSearchResultTitleTokens := tokeniseTitle(searchResult.Title)
			goodReadsSearchBookTitleTokens := tokeniseTitle(bookInfo.Title)

			if titlesMatch(goodReadsSearchBookTitleTokens, theBookSearchResultTitleTokens) {
				potentialTitleMatches = append(potentialTitleMatches, searchResult)
			}
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
	hasNonEnglishCharacters := NON_ENGLISH_CHARACTER.MatchString(bookDetail)
	if hasNonEnglishCharacters {
		return false
	}

	// A very crude BUT lightweight way of detecting a good amount
	// of non english books. Using an actual language detection
	// library would be like using an airplane to thread a needle
	experimentalNonEnglishSnippets := []string{
		" de ",
		" le ",
		" en ",
		" francais ",
		" del ",
		" el ",
		" los ",
		" las ",
		" und ",
		" der ",
		" des ",
		" dem ",
		" y ",
		" ein ",
		" eine ",
		" einer ",
		" l'",
		" d'",
		" la ",
		" c'est ",
	}
	for _, nonEnglishSnippet := range experimentalNonEnglishSnippets {
		if strings.Contains(strings.ToLower(bookDetail), nonEnglishSnippet) {
			return false
		}
	}

	return true
}

func getPureTheBookshopTitle(unfilteredTitle string) string {
	_, titleWithoutParenthesesText := removeUnnecessaryBitsFromTheBookshopTitle(unfilteredTitle)
	titleWithoutDashesText := removeAllTextAfterFirstDashIfPossible(titleWithoutParenthesesText)
	return titleWithoutDashesText
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

func getAuthorFullNameInCorrectOrder(authorWithLastnameFirst string) string {
	nameArr := strings.Split(authorWithLastnameFirst, ",")
	if len(nameArr) == 1 {
		return strings.TrimSpace(nameArr[0])
	}

	nameArr = removeEmptyStringElementsInArr(nameArr)
	if len(nameArr) == 2 {
		return fmt.Sprintf("%s %s", strings.TrimSpace(nameArr[1]), strings.TrimSpace(nameArr[0]))
	}

	logger.Sugar().Infof("Failed to split author '%s' by non existant comma", authorWithLastnameFirst)
	return authorWithLastnameFirst
}

func tokeniseTitle(title string) []string {
	titleWords := strings.Split(getPureTheBookshopTitle(title), " ")
	titleWordsNoEmpties := removeEmptyStringElementsInArr(titleWords)
	return lowercaseAllStringElements(titleWordsNoEmpties)
}

func removeEmptyStringElementsInArr(arr []string) []string {
	noEmptyElementsArr := []string{}
	for _, elem := range arr {
		if elem != "" {
			noEmptyElementsArr = append(noEmptyElementsArr, elem)
		}
	}
	return noEmptyElementsArr
}

func lowercaseAllStringElements(arr []string) []string {
	lowerCasedArr := []string{}
	for _, elem := range arr {
		lowerCasedArr = append(lowerCasedArr, strings.ToLower(elem))
	}
	return lowerCasedArr
}

func titlesMatch(searchBookTokens, searchResultTokens []string) bool {
	if len(searchBookTokens) != len(searchResultTokens) {
		return false
	}
	for i, searchBookToken := range searchBookTokens {
		if searchBookToken != searchResultTokens[i] {
			return false
		}
	}
	return true
}
