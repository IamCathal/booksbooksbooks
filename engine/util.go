package engine

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func writeTotalBooksMsg(stats dtos.CrawlStats, ws *websocket.Conn) {
	totalBooksMsg := dtos.WsTotalBooks{
		TotalBooks: stats.TotalBooks,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(totalBooksMsg)
	if err != nil {
		panic(err)
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
		panic(err)
	}
	WriteMsg(jsonStr, ws)
}

func writeSearchResultReturnedMsg(searchResult dtos.AllBookshopBooksSearchResults, stats dtos.CrawlStats, ws *websocket.Conn) {
	searchResultMsg := dtos.WsBookshopSearchResult{
		Result:     searchResult,
		CrawlStats: stats,
	}
	jsonStr, err := json.Marshal(searchResultMsg)
	if err != nil {
		panic(err)
	}
	WriteMsg(jsonStr, ws)
}

func WriteMsg(msg []byte, ws *websocket.Conn) {
	err := ws.WriteMessage(1, msg)
	if err != nil {
		panic(err)
	}
}
