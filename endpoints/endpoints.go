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
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
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
	r.HandleFunc("/automatedcheck", automatedCheck).Methods("POST")
	r.HandleFunc("/getavailablebooks", getAvailableBooks).Methods("GET")
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

func automatedCheck(w http.ResponseWriter, r *http.Request) {
	stubStatsChan := make(chan int, 1)
	stubBooksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	stubSearchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

	cachedBooksThatWereAvailable := db.GetAvailableBooks()
	cachedBooksThatAreStillAvailableToday := []dtos.AvailableBook{}
	booksFromShelfThatAreAvailable := []dtos.AvailableBook{}

	for _, book := range cachedBooksThatWereAvailable {
		searchResult := thebookshop.SearchForBook(book.BookInfo, stubSearchResultsFromTheBookshopChan)

		if len(searchResult.TitleMatches) >= 1 {
			cachedBooksThatAreStillAvailableToday = append(cachedBooksThatAreStillAvailableToday, book)
		}
	}

	fmt.Printf("Previously available books: %d\n", len(cachedBooksThatWereAvailable))
	fmt.Printf("Now available books %d\n", len(cachedBooksThatAreStillAvailableToday))

	shelfURL := "https://www.goodreads.com/review/list/151819645-cathal?ref=nav_mybooks&shelf=yet-to-read"
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfURL '%s' given", shelfURL)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	fmt.Printf("shelfURL is valid\n")

	booksFromShelf := goodreads.GetBooksFromShelf(shelfURL, stubStatsChan, stubBooksFoundFromGoodReadsChan)
	close(stubBooksFoundFromGoodReadsChan)

	fmt.Printf("Got back %d books\n", len(booksFromShelf))
	searchResults := []dtos.EnchancedSearchResult{}

	for _, book := range booksFromShelf {
		searchResults = append(searchResults, thebookshop.SearchForBook(book, stubSearchResultsFromTheBookshopChan))
	}

	fmt.Printf("Here are all of the search results:\n")
	for i, res := range searchResults {
		fmt.Printf("[%d] %+v\n", i, res)
	}
	booksFromShelfThatAreAvailable = goodreads.GetAvailableBooksFromSearchResult(searchResults)

	newBooksThatNeedNotification := []dtos.AvailableBook{}
	for _, availableBook := range booksFromShelfThatAreAvailable {
		if bookIsNew := bookIsNew(availableBook, cachedBooksThatAreStillAvailableToday); bookIsNew {
			newBooksThatNeedNotification = append(newBooksThatNeedNotification, availableBook)
		}
	}

	if len(newBooksThatNeedNotification) > 0 {
		for _, newBook := range newBooksThatNeedNotification {
			db.AddAvailableBook(newBook)
		}
	}
	fmt.Printf("%d books were available yesterday\n", len(cachedBooksThatWereAvailable))
	fmt.Printf("%d books are available today from cache\n", len(cachedBooksThatAreStillAvailableToday))
	fmt.Printf("These books are brand new from this current crawl: %+v\n", newBooksThatNeedNotification)

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
