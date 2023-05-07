package endpoints

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/engine"
)

func setupWebsocketsRouter(mainRouter *mux.Router) {
	websocketRouter := mainRouter.PathPrefix("/ws").Subrouter()
	websocketRouter.HandleFunc("/shelfcrawl", shelfCrawl).Methods("GET")
	websocketRouter.HandleFunc("/seriescrawl", seriesCrawl).Methods("GET")
	websocketRouter.Use(logMiddleware)
}

func shelfCrawl(w http.ResponseWriter, r *http.Request) {
	ws := setupWebSocket(w, r)
	if ws == nil {
		SendBasicInvalidResponse(w, r, "unable to upgrade websocket", http.StatusBadRequest)
		return
	}
	engine.GenericWorker(db.GetShelfURLsFromShelvesToCrawl(), ws)
	ws.Close()
}

func seriesCrawl(w http.ResponseWriter, r *http.Request) {
	ws := setupWebSocket(w, r)
	if ws == nil {
		SendBasicInvalidResponse(w, r, "unable to upgrade websocket", http.StatusBadRequest)
		return
	}
	engine.SeriesLookupWorker(ws)
	ws.Close()
}
