package endpoints

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/engine"
)

var (
	appConfig dtos.AppConfig
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func InitConfig(conf dtos.AppConfig) {
	appConfig = conf
}

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/available", available).Methods("GET")
	r.HandleFunc("/status", status).Methods("POST")
	r.HandleFunc("/recentcrawls", getRecentCrawls).Methods("GET")
	// r.HandleFunc("/automatedcheck", automaticCheck).Methods("POST")
	r.HandleFunc("/ws", liveFeed).Methods("GET")
	r.Use(logMiddleware)

	r.Handle("/static", http.NotFoundHandler())
	fs := http.FileServer(http.Dir(""))
	r.PathPrefix("/").Handler(DisallowFileBrowsing(fs))
	return r
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/index.html")
}

func available(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/available.html")
}

// func automaticCheck(w http.ResponseWriter, r *http.Request) {
// 	previouslyAvailableBooks := db.GetAvailableBooks()
// 	booksThatAreStillAvailable := []dtos.TheBookshopBook{}
// 	searchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

// 	for _, book := range previouslyAvailableBooks {
// 		searchResult := thebookshop.SearchForBook(book, searchResultsFromTheBookshopChan)
// 	}
// }

func liveFeed(w http.ResponseWriter, r *http.Request) {
	ws := setupWebSocket(w, r)
	if ws == nil {
		SendBasicInvalidResponse(w, r, "unable to upgrade websocket", http.StatusBadRequest)
		return
	}
	engine.Worker(r.URL.Query().Get("shelfurl"), ws)
	ws.Close()
}

func getRecentCrawls(w http.ResponseWriter, r *http.Request) {
	recentCrawls := db.GetRecentCrawls()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(recentCrawls)
}

func status(w http.ResponseWriter, r *http.Request) {
	req := dtos.UptimeResponse{
		Status:      "operational",
		Uptime:      time.Duration(time.Since(appConfig.ApplicationStartUpTime).Milliseconds()),
		StartUpTime: appConfig.ApplicationStartUpTime.Unix(),
	}
	jsonObj, err := json.MarshalIndent(req, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(string(jsonObj))
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isActualEndpoint := isActualEndpoint(r.URL.Path); isActualEndpoint {
			fmt.Printf("%v %+v\n", time.Now().Format(time.RFC3339), r)
		}
		next.ServeHTTP(w, r)
	})
}
