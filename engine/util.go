package engine

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/dtos"
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

func bookIsNew(book dtos.TheBookshopBook, availableBooksMap map[string]bool) bool {
	_, exists := availableBooksMap[book.Link]
	return !exists
}
