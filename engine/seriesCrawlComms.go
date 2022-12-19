package engine

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func writeOverallSeriesCrawlStatsMessage(crawlStats dtos.SeriesCrawlStats, ws *websocket.Conn) {
	crawlStatsMsg := dtos.WsSeriesCrawlStats{
		CrawlStats: crawlStats,
	}
	jsonStr, err := json.Marshal(crawlStatsMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func writeNewSeriesFoundMessage(newSeries dtos.Series, crawlStats dtos.SeriesCrawlStats, ws *websocket.Conn) {
	crawlStatsMsg := dtos.WsNewSeries{
		Series:     newSeries,
		CrawlStats: crawlStats,
	}
	jsonStr, err := json.Marshal(crawlStatsMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func writeSearchResultReturnedMessage(searchBook dtos.BasicGoodReadsBook, match dtos.TheBookshopBook, crawlStats dtos.SeriesCrawlStats, ws *websocket.Conn) {
	crawlStatsMsg := dtos.WsSearchResultReturned{
		SearchBook: searchBook,
		Match:      match,
		CrawlStats: crawlStats,
	}
	jsonStr, err := json.Marshal(crawlStatsMsg)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = controller.Cnt.WriteWsMessage(jsonStr, ws)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}
