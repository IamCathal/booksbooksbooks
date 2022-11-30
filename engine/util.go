package engine

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/util"
)

var (
	websocketWriteLock sync.Mutex
)

func writeErrorMsg(msg string, ws *websocket.Conn) {
	errorMsg := dtos.WsErrorMsg{
		Error: msg,
	}
	jsonStr, err := json.Marshal(errorMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	WriteMsg(jsonStr, ws)
}

func writeTotalBooksMsg(stats dtos.CrawlStats, ws *websocket.Conn) {
	totalBooksMsg := dtos.WsTotalBooks{
		TotalBooks: stats.TotalBooks,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(totalBooksMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	WriteMsg(jsonStr, ws)
}

func writeGoodReadsBookMsg(bookInfo dtos.BasicGoodReadsBook, stats dtos.CrawlStats, ws *websocket.Conn) {
	goodReadsBookMsg := dtos.WsGoodreadsBook{
		BookInfo:   bookInfo,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(goodReadsBookMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	WriteMsg(jsonStr, ws)
}

func writeNewAvailableBookMsg(bookInfo dtos.TheBookshopBook, stats dtos.CrawlStats, ws *websocket.Conn) {
	newAvaialbleBookMsg := dtos.WsNewBookAvailable{
		Book:       bookInfo,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(newAvaialbleBookMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	WriteMsg(jsonStr, ws)
}

func writeSearchResultReturnedMsg(searchResult dtos.EnchancedSearchResult, stats dtos.CrawlStats, ws *websocket.Conn) {
	searchResultMsg := dtos.WsBookshopSearchResult{
		SearchResult: searchResult,
		CrawlStats:   stats,
	}
	jsonStr, err := json.Marshal(searchResultMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	WriteMsg(jsonStr, ws)
}

func WriteMsg(msg []byte, ws *websocket.Conn) {
	websocketWriteLock.Lock()
	defer websocketWriteLock.Unlock()
	err := ws.WriteMessage(1, msg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func goodReadsBookIsNew(book dtos.TheBookshopBook, availableBooksMap map[string]bool) bool {
	_, exists := availableBooksMap[book.Link]
	return !exists
}

func getFormattedTime() string {
	now := time.Now()

	currHour := now.Hour()
	currHourFormatted := ""
	if currHour >= 0 && currHour <= 9 {
		currHourFormatted = fmt.Sprintf("0%d", currHour)
	} else {
		currHourFormatted = fmt.Sprint(now.Hour())
	}

	currMinute := now.Minute()
	currMinuteFormatted := ""
	if currMinute >= 0 && currMinute <= 9 {
		currMinuteFormatted += fmt.Sprintf("0%d", currMinute)
	} else {
		currMinuteFormatted = fmt.Sprint(now.Minute())
	}

	return fmt.Sprintf("%s:%s", currHourFormatted, currMinuteFormatted)
}

func sendFreeShippingWebhookIfFreeShippingEligible() {
	allAvailableBooks := db.GetAvailableBooks()
	var totalCost float64
	for _, book := range allAvailableBooks {
		if !book.Ignore {
			totalCost += extractFloatPriceFromString(book.BookPurchaseInfo.Price)
		}
	}
	if totalCost >= 20 {
		util.SendFreeShippingTotalHasKickedInMessage(totalCost)
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
