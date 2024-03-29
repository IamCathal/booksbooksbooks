package search

import (
	"fmt"
	"strings"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/util"
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
		searchResultAuthor := getAuthorFullNameInCorrectOrder(searchResult.Author)
		searchBookAuthor := bookInfo.Author
		if isAuthorNameReversed(bookInfo.Author) {
			// When retrieving author names for books on a series page the author's
			// names are in correct order e.g Patrick Rothfuss
			// When retrieving author names from a shelf page the author's
			// name are in reverse order e.g Rothfuss, Patrick
			searchBookAuthor = getAuthorFullNameInCorrectOrder(bookInfo.Author)
		}

		if strings.EqualFold(searchBookAuthor, searchResultAuthor) {
			potentialAuthorMatches = append(potentialAuthorMatches, searchResult)

			if strings.Contains(searchResult.Title, ":") {
				continue
			}

			theBookshopPureTitle := getPureTheBookshopTitle(searchResult.Title)
			theBookSearchResultTitleTokens := tokeniseTitle(theBookshopPureTitle)
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
		logger.Sugar().Infof("%d potential title matches found for book: %+v, title matches: %+v",
			len(potentialTitleMatches), bookInfo.Title, util.GetConciseInfoFromGoodReadsBooks(potentialTitleMatches))
	}
	if len(potentialAuthorMatches) >= 1 {
		logger.Sugar().Infof("%d potential author matches found for author %s who wrote %+v, author matches: %+v",
			len(potentialAuthorMatches), bookInfo.Author, bookInfo.Title, util.GetConciseInfoFromGoodReadsBooks(potentialAuthorMatches))
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
		isAuthorEnglish := util.IsEnglishText(titleMatch.Author)
		isTitleEnglish := util.IsEnglishText(titleMatch.Title)

		if isAuthorEnglish && isTitleEnglish {
			filteredSearchResults.TitleMatches = append(filteredSearchResults.TitleMatches, titleMatch)
		}
	}
	for _, authorMatch := range searchResult.AuthorMatches {
		isAuthorEnglish := util.IsEnglishText(authorMatch.Author)
		isTitleEnglish := util.IsEnglishText(authorMatch.Title)

		if isAuthorEnglish && isTitleEnglish {
			filteredSearchResults.AuthorMatches = append(filteredSearchResults.AuthorMatches, authorMatch)
		} else {
			// logger.Sugar().Infof("Filtering out non-english book: %s by %s because filter non-english books is enabled", authorMatch.Title, authorMatch.Author)
		}
	}

	return filteredSearchResults
}

func getPureTheBookshopTitle(unfilteredTitle string) string {
	titleWithoutParenthesesText := removeUnnecessaryBitsFromTheBookshopTitle(unfilteredTitle)
	titleWithoutDashesText := removeAllTextAfterFirstDashIfPossible(titleWithoutParenthesesText)
	return titleWithoutDashesText
}

func removeUnnecessaryBitsFromTheBookshopTitle(fullTitleText string) string {
	titleWithoutParethesisText := removeAllBetweenSubStrings(fullTitleText, "(", ")")
	titleWithoutDashesText := removeAllTextAfterFirstDashIfPossible(titleWithoutParethesisText)

	return strings.TrimSpace(titleWithoutDashesText)
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

	logger.Sugar().Infof("Failed to split author '%s' by non existant or >3 commas", authorWithLastnameFirst)
	return authorWithLastnameFirst
}

func isAuthorNameReversed(authorString string) bool {
	// this is very experimental
	return strings.ContainsAny(authorString, ",")
}

func removeCommaFromAuthorName(authorName string) string {
	return strings.ReplaceAll(authorName, ",", "")
}

func tokeniseTitle(title string) []string {
	titleWords := strings.Split(title, " ")
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
