package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
	"github.com/iamcathal/booksbooksbooks/util"
)

func writeErrorMsg(msg string, ws *websocket.Conn) {
	errorMsg := dtos.WsErrorMsg{
		Error: msg,
	}
	jsonStr, err := json.Marshal(errorMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func checkAvailabilityOfExistingAvailableBooksList(booksThatWereAvailableLastTime []dtos.AvailableBook) {
	logger.Sugar().Infof("%d books were available from the last automated check: %v",
		len(booksThatWereAvailableLastTime), getConciseBookInfoFromAvailableBooks(booksThatWereAvailableLastTime))

	booksFromLastTimeThatAreStillAvailable := lookUpAvailabilityOfBooksThatWerePreviouslyAvailable(booksThatWereAvailableLastTime)
	logger.Sugar().Infof("%d books from the previous available list that are still available now: %v",
		len(booksFromLastTimeThatAreStillAvailable), getConciseBookInfoFromAvailableBooks(booksFromLastTimeThatAreStillAvailable))

	booksThatAreNowNotAvailable := findBooksThatAreNowNotAvailable(booksThatWereAvailableLastTime, booksFromLastTimeThatAreStillAvailable)
	for _, book := range booksThatAreNowNotAvailable {
		db.RemoveAvailableBook(book)
		if alertOnNoLongerAvailableBooks := db.GetSendAlertWhenBookNoLongerAvailable(); alertOnNoLongerAvailableBooks {
			util.SendNewBookIsAvailableNotification(book.BookPurchaseInfo, false)
		}
	}
}

func lookUpAvailabilityOfBooksThatWerePreviouslyAvailable(previouslyAvailableBooks []dtos.AvailableBook) []dtos.AvailableBook {
	stubSearchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)
	booksThatAreStillAvailable := []dtos.AvailableBook{}

	for _, book := range previouslyAvailableBooks {
		searchBook := dtos.BasicGoodReadsBook{
			Title:  book.BookPurchaseInfo.Title,
			Author: book.BookPurchaseInfo.Author,
		}
		logger.Sugar().Infof("Checking availability of previously available book: %s by %s (found through %d)",
			searchBook.Author, searchBook.Title, book.BookFoundFrom)
		searchResult := thebookshop.SearchForBook(searchBook, stubSearchResultsFromTheBookshopChan)

		if len(searchResult.TitleMatches) > 0 {
			currAvailableBook := dtos.AvailableBook{
				BookInfo:         book.BookInfo,
				BookPurchaseInfo: searchResult.TitleMatches[0],
				Ignore:           book.Ignore,
			}
			booksThatAreStillAvailable = append(booksThatAreStillAvailable, currAvailableBook)
		}

	}

	return booksThatAreStillAvailable
}

func wasNotPreviouslyAvailable(book dtos.TheBookshopBook, availableBooksMap map[string]bool) bool {
	_, exists := availableBooksMap[book.Link]
	return !exists
}

func sendFreeShippingWebhookIfFreeShippingEligible() {
	allAvailableBooks := db.GetAvailableBooks()
	var totalCost float64

	for _, book := range allAvailableBooks {
		if !book.Ignore {
			totalCost += extractFloatPriceFromString(book.BookPurchaseInfo.Price)
		}
	}
	if totalCost >= FREE_SHIPPING_THRESHOLD {
		util.SendFreeShippingTotalHasKickedInNotification(totalCost)
	}
}

func extractFloatPriceFromString(priceString string) float64 {
	stringPriceNoEuroSymbol := strings.ReplaceAll(priceString, "€", "")
	floatPrice, err := strconv.ParseFloat(stringPriceNoEuroSymbol, 64)
	if err != nil {
		panic(err)
	}
	return floatPrice
}

func notifyAboutBooksThatAreNoLongerAvailable(previouslyAvailable []dtos.AvailableBook) {
	currAvailableMap := db.GetAvailableBooksMap()
	updatedCurrentlyAvailableBooks := []dtos.AvailableBook{}

	for _, prevAvailableBook := range previouslyAvailable {
		if _, isStillAvailable := currAvailableMap[prevAvailableBook.BookPurchaseInfo.Link]; !isStillAvailable {
			logger.Sugar().Infof("Removing book %s - %s (%s) as its no longer available",
				prevAvailableBook.BookInfo.Author,
				prevAvailableBook.BookInfo.Title,
				prevAvailableBook.BookPurchaseInfo.Link)
			util.SendBookIsNoLongerAvailableNotification(prevAvailableBook.BookPurchaseInfo)
		} else {
			updatedCurrentlyAvailableBooks = append(updatedCurrentlyAvailableBooks, prevAvailableBook)
		}
	}
	db.SetAvailableBooks(updatedCurrentlyAvailableBooks)
}

func filterIgnoredAuthors(searchResult dtos.EnchancedSearchResult) dtos.EnchancedSearchResult {
	filteredSearchResult := dtos.EnchancedSearchResult{
		SearchBook: searchResult.SearchBook,
	}

	for _, titleMatch := range searchResult.TitleMatches {
		if authorIsIgnored := db.IsIgnoredAuthor(titleMatch.Author); !authorIsIgnored {
			filteredSearchResult.TitleMatches = append(filteredSearchResult.TitleMatches, titleMatch)
		}
	}

	for _, authorMatches := range searchResult.AuthorMatches {
		if authorIsIgnored := db.IsIgnoredAuthor(authorMatches.Author); !authorIsIgnored {
			filteredSearchResult.AuthorMatches = append(filteredSearchResult.AuthorMatches, authorMatches)
		}
	}

	return filteredSearchResult
}

func findBooksThatAreNowNotAvailable(availableThen, availableNow []dtos.AvailableBook) []dtos.AvailableBook {
	booksThatAreNoLongerAvailable := []dtos.AvailableBook{}
	availableNowMap := make(map[string]bool)

	for _, book := range availableNow {
		availableNowMap[book.BookInfo.ID] = true
	}

	for _, book := range availableThen {
		if _, exists := availableNowMap[book.BookInfo.ID]; !exists {
			booksThatAreNoLongerAvailable = append(booksThatAreNoLongerAvailable, book)
		}
	}

	return booksThatAreNoLongerAvailable
}

func getConciseBookInfoFromAvailableBooks(bookList []dtos.AvailableBook) []string {
	info := []string{}

	for _, book := range bookList {
		info = append(info, fmt.Sprintf("%s: %s", book.BookInfo.Author, book.BookInfo.Title))
	}
	return info
}

func extractGoodreadsBooksThatAreInSeries(allBooks []dtos.BasicGoodReadsBook) []dtos.BasicGoodReadsBook {
	booksThatAreInASeries := []dtos.BasicGoodReadsBook{}
	for _, book := range allBooks {
		if book.SeriesText != "" {
			booksThatAreInASeries = append(booksThatAreInASeries, book)
		}
	}
	return booksThatAreInASeries
}

func extractAvailableBooksThatAreInSeries(allBooks []dtos.AvailableBook) []dtos.AvailableBook {
	booksThatAreInASeries := []dtos.AvailableBook{}
	for _, book := range allBooks {
		if book.BookInfo.SeriesText != "" {
			booksThatAreInASeries = append(booksThatAreInASeries, book)
		}
	}
	return booksThatAreInASeries
}

func mergeBooksThatAreInASeries(ownedBooks []dtos.BasicGoodReadsBook, availableBooks []dtos.AvailableBook) []dtos.AvailableBook {
	mergedBooksList := []dtos.AvailableBook{}
	seenBooks := make(map[string]bool)

	for _, ownedBook := range ownedBooks {
		key := fmt.Sprintf("%s/%s", ownedBook.Author, ownedBook.Title)
		if _, exists := seenBooks[key]; !exists {
			seenBooks[key] = true
			availableBook := dtos.AvailableBook{
				BookInfo: ownedBook,
			}
			mergedBooksList = append(mergedBooksList, availableBook)
		}
	}

	for _, availableBook := range availableBooks {
		key := fmt.Sprintf("%s/%s", availableBook.BookInfo.Author, availableBook.BookInfo.Title)
		if _, exists := seenBooks[key]; !exists {
			seenBooks[key] = true
			mergedBooksList = append(mergedBooksList, availableBook)
		}
	}

	return mergedBooksList
}
