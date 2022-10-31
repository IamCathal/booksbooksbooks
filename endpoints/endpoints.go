package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/engine"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
	"go.uber.org/zap"
)

var (
	logger    *zap.Logger
	appConfig dtos.AppConfig
	upgrader  = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func InitConfig(conf dtos.AppConfig, newLogger *zap.Logger) {
	appConfig = conf
	logger = newLogger
}

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/ws", liveFeed).Methods("GET")
	r.HandleFunc("/status", status).Methods("POST")
	r.HandleFunc("/available", available).Methods("GET")
	r.HandleFunc("/getrecentcrawls", getRecentCrawls).Methods("GET")
	r.HandleFunc("/automatedcheck", automatedCheck).Methods("POST")
	r.HandleFunc("/getavailablebooks", getAvailableBooks).Methods("GET")
	r.HandleFunc("/resetavailablebooks", resetAvailableBooks).Methods("POST")
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

func automatedCheck(w http.ResponseWriter, r *http.Request) {
	stubStatsChan := make(chan int, 1)
	stubBooksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	stubSearchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

	cachedBooksThatWereAvailable := db.GetAvailableBooks()
	cachedBooksThatAreStillAvailableToday := []dtos.AvailableBook{}
	booksFromShelfThatAreAvailableNow := []dtos.AvailableBook{}

	for _, book := range cachedBooksThatWereAvailable {
		searchResult := thebookshop.SearchForBook(book.BookInfo, stubSearchResultsFromTheBookshopChan)

		if len(searchResult.TitleMatches) >= 1 {
			cachedBooksThatAreStillAvailableToday = append(cachedBooksThatAreStillAvailableToday, book)
		}
	}

	logger.Sugar().Infof("%d cached books that were available from the last automated checkup: %d\n",
		len(cachedBooksThatWereAvailable), cachedBooksThatWereAvailable)
	logger.Sugar().Infof("%d Cached froom from the last automated checkup that are still available now: %d\n",
		len(cachedBooksThatAreStillAvailableToday), cachedBooksThatAreStillAvailableToday)

	shelfURL := "https://www.goodreads.com/review/list/151819645-cathal?ref=nav_mybooks&shelf=yet-to-read"
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfURL '%s' given", shelfURL)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}

	booksFromShelf := goodreads.GetBooksFromShelf(shelfURL, stubStatsChan, stubBooksFoundFromGoodReadsChan)
	logger.Sugar().Infof("%d books were found from GoodReads shelf %s\n", len(booksFromShelf), shelfURL)
	close(stubBooksFoundFromGoodReadsChan)

	searchResults := []dtos.EnchancedSearchResult{}
	for _, book := range booksFromShelf {
		searchResults = append(searchResults, thebookshop.SearchForBook(book, stubSearchResultsFromTheBookshopChan))
	}
	booksFromShelfThatAreAvailableNow = goodreads.GetAvailableBooksFromSearchResult(searchResults)
	logger.Sugar().Infof("%s search queries were made with %d matches found",
		len(searchResults), len(booksFromShelfThatAreAvailableNow))

	newBooksThatNeedNotification := []dtos.AvailableBook{}
	for _, availableBook := range booksFromShelfThatAreAvailableNow {
		if bookIsNew := bookIsNew(availableBook, cachedBooksThatAreStillAvailableToday); bookIsNew {
			newBooksThatNeedNotification = append(newBooksThatNeedNotification, availableBook)
		}
	}

	logger.Sugar().Infof("%d new books were found in this search", len(newBooksThatNeedNotification))
	if len(newBooksThatNeedNotification) > 0 {
		for _, newBook := range newBooksThatNeedNotification {
			db.AddAvailableBook(newBook)
		}
	}
	logger.Sugar().Infof("%d cached books were available yesterday", len(cachedBooksThatWereAvailable))
	logger.Sugar().Infof("%d books are available today from cache", len(cachedBooksThatAreStillAvailableToday))
	logger.Sugar().Infof("These books are brand new from this current crawl: %+v\n", newBooksThatNeedNotification)

	fmt.Fprintf(w, "hello world")
}

func liveFeed(w http.ResponseWriter, r *http.Request) {
	ws := setupWebSocket(w, r)
	if ws == nil {
		SendBasicInvalidResponse(w, r, "unable to upgrade websocket", http.StatusBadRequest)
		return
	}
	engine.Worker(r.URL.Query().Get("shelfurl"), ws)
	ws.Close()
}

func getAvailableBooks(w http.ResponseWriter, r *http.Request) {
	recentCrawls := db.GetAvailableBooks()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(recentCrawls)
}

func resetAvailableBooks(w http.ResponseWriter, r *http.Request) {
	db.ResetAvailableBooks()
	w.WriteHeader(http.StatusOK)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req)
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isActualEndpoint := isActualEndpoint(r.URL.Path); isActualEndpoint {
			logger.Sugar().Infow(fmt.Sprintf("Served request to %v", r.URL.Path),
				zap.String("requestInfo", fmt.Sprintf("%+v", r)))
		}
		next.ServeHTTP(w, r)
	})
}
