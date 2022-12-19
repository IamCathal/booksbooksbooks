package engine

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func writeTotalBooksInShelfWsMessage(stats dtos.CrawlStats, ws *websocket.Conn) {
	totalBooksMsg := dtos.WsTotalBooks{
		TotalBooks: stats.TotalBooks,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(totalBooksMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func writeBookFromShelfWsMessage(bookInfo dtos.BasicGoodReadsBook, stats dtos.CrawlStats, ws *websocket.Conn) {
	goodReadsBookMsg := dtos.WsGoodreadsBook{
		BookInfo:   bookInfo,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(goodReadsBookMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func writeNewAvailableBookWsMsg(bookInfo dtos.TheBookshopBook, stats dtos.CrawlStats, ws *websocket.Conn) {
	newAvaialbleBookMsg := dtos.WsNewBookAvailable{
		Book:       bookInfo,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(newAvaialbleBookMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
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
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}
