package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/engine"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/util"
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
	r.HandleFunc("/settings", settings).Methods("GET")
	r.HandleFunc("/available", available).Methods("GET")
	r.HandleFunc("/getrecentcrawls", getRecentCrawls).Methods("GET")
	r.HandleFunc("/getavailablebooks", getAvailableBooks).Methods("GET")
	r.HandleFunc("/ignorebook", ignoreBook).Methods("POST")
	r.HandleFunc("/unignorebook", unignoreBook).Methods("POST")
	r.HandleFunc("/resetavailablebooks", resetAvailableBooks).Methods("POST")
	r.HandleFunc("/getautomatedcrawlshelfstats", getAutomatedCrawlShelfStats).Methods("GET")
	r.Use(logMiddleware)

	settingsRouter := r.PathPrefix("/settings").Subrouter()
	settingsRouter.HandleFunc("/getpreviewforshelf", getPreviewForShelf).Methods("GET")
	settingsRouter.HandleFunc("/setdiscordmessageformat", setDiscordMessageFormat).Methods("POST")
	settingsRouter.HandleFunc("/getdiscordmessageformat", getDiscordMessageFormat).Methods("GET")
	settingsRouter.HandleFunc("/setautomatedcrawltime", setAutomatedCrawlTime).Methods("POST")
	settingsRouter.HandleFunc("/getautomatedcrawltime", getAutomatedCrawlTime).Methods("GET")
	settingsRouter.HandleFunc("/getautomatedbookshelfcheckurl", getautomatedbookshelfcheckurl).Methods("GET")
	settingsRouter.HandleFunc("/setautomatedbookshelfcheckurl", setautomatedbookshelfcheckurl).Methods("POST")
	settingsRouter.HandleFunc("/testdiscordwebhook", testDiscordWebook).Methods("GET")
	settingsRouter.HandleFunc("/setdiscordwebhook", setDiscordWebook).Methods("POST")
	settingsRouter.HandleFunc("/getdiscordwebhook", getDiscordWebook).Methods("GET")
	settingsRouter.HandleFunc("/setsendalertwhenbooknolongeravailable", setSendAlertWhenBookNoLongerAvailable).Methods("POST")
	settingsRouter.HandleFunc("/getsendalertwhenbooknolongeravailable", getSendAlertWhenBookNoLongerAvailable).Methods("GET")
	settingsRouter.HandleFunc("/setsendalertonlywhenfreeshippingkicksin", setSendAlertOnlyWhenFreeShippingKicksIn).Methods("POST")
	settingsRouter.HandleFunc("/getsendalertonlywhenfreeshippingkicksin", getSendAlertOnlyWhenFreeShippingKicksIn).Methods("GET")
	settingsRouter.Use(logMiddleware)

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

func settings(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/settings.html")
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
	availableBooks := db.GetAvailableBooks()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(availableBooks)
}

func ignoreBook(w http.ResponseWriter, r *http.Request) {
	bookURL := r.URL.Query().Get("bookurl")
	db.IgnoreBook(bookURL)
	w.WriteHeader(http.StatusOK)
}

func unignoreBook(w http.ResponseWriter, r *http.Request) {
	bookURL := r.URL.Query().Get("bookurl")
	db.UnignoreBook(bookURL)
	w.WriteHeader(http.StatusOK)
}

func testDiscordWebook(w http.ResponseWriter, r *http.Request) {
	discordWebhook := r.URL.Query().Get("webhookurl")
	db.SetDiscordWebhookURL(discordWebhook)
	util.DeliverWebHook(dtos.DiscordMsg{
		Username:   "BooksBooksBooks",
		Avatar_url: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
		Embed: []dtos.DiscordEmbed{
			{
				Title: "BooksBooksBooks is ready to send webhook updates to this channel",
			},
		},
	})
	db.SetDiscordWebhookURL(discordWebhook)
	w.WriteHeader(http.StatusOK)
}

func setDiscordWebook(w http.ResponseWriter, r *http.Request) {
	discordWebhook := r.URL.Query().Get("webhookurl")
	db.SetDiscordWebhookURL(discordWebhook)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func getDiscordWebook(w http.ResponseWriter, r *http.Request) {
	res := dtos.GetDiscordWebhookResponse{
		WebHook: db.GetDiscordWebhookURL(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func resetAvailableBooks(w http.ResponseWriter, r *http.Request) {
	db.ResetAvailableBooks()
	w.WriteHeader(http.StatusOK)
}

func getAutomatedCrawlShelfStats(w http.ResponseWriter, r *http.Request) {
	shelfURL := db.GetAutomatedBookShelfCheck()
	nonIgnoredBookCount, ignoredBookCount := db.GetIgnoredAndNonIgnoredCountOfAvailableBooks()

	res := dtos.GetAutomatedCrawlShelfStats{
		ShelfBreadcrumb:       db.GetKeyForRecentCrawlBreadcrumb(shelfURL),
		ShelfURL:              shelfURL,
		TotalBooks:            db.GetTotalBooksInAutomatedBookShelfCheck(),
		AvailableBooks:        nonIgnoredBookCount,
		IgnoredAvailableBooks: ignoredBookCount,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func getPreviewForShelf(w http.ResponseWriter, r *http.Request) {
	shelfURL := r.URL.Query().Get("shelfurl")
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfurl '%s' given", shelfURL)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}

	books, totalBooks := goodreads.GetPreviewForShelf(shelfURL)
	res := dtos.GetPreviewForShelfResponse{
		Books:      books,
		TotalBooks: totalBooks,
	}
	db.SetTotalBooksInAutomatedBookShelfCheck(totalBooks)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func getRecentCrawls(w http.ResponseWriter, r *http.Request) {
	recentCrawls := db.GetRecentCrawlBreadcrumbs()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(recentCrawls)
}

func getautomatedbookshelfcheckurl(w http.ResponseWriter, r *http.Request) {
	bookShelfURL := db.GetAutomatedBookShelfCheck()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dtos.AutomatedShelfCheckURLResponse{ShelURL: bookShelfURL})
}

func setautomatedbookshelfcheckurl(w http.ResponseWriter, r *http.Request) {
	bookShelfURL := r.URL.Query().Get("shelfurl")
	if isValidShelfURL := goodreads.CheckIsShelfURL(bookShelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfurl '%s' given", bookShelfURL)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}

	db.SetAutomatedBookShelfCheck(bookShelfURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookShelfURL)
}

func getDiscordMessageFormat(w http.ResponseWriter, r *http.Request) {
	res := dtos.GetDiscordMessageFormatResponse{
		Format: db.GetDiscordMessageFormat(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func setDiscordMessageFormat(w http.ResponseWriter, r *http.Request) {
	messageFormat := r.URL.Query().Get("messageformat")
	if messageFormat != "big" && messageFormat != "small" {
		errorMsg := fmt.Sprintf("Invalid message format '%s' given", messageFormat)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetDiscordMessageFormat(messageFormat)
	w.WriteHeader(http.StatusOK)
}

func setAutomatedCrawlTime(w http.ResponseWriter, r *http.Request) {
	automatedCrawlTime := r.URL.Query().Get("time")
	_, err := time.Parse("15:04:05", fmt.Sprintf("%s:00", automatedCrawlTime))
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid crawl time '%s' given. must be in format hh:mm", automatedCrawlTime)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetAutomatedBookShelfCrawlTime(automatedCrawlTime)
	w.WriteHeader(http.StatusOK)
}

func getAutomatedCrawlTime(w http.ResponseWriter, r *http.Request) {
	res := dtos.GetAutomatedCrawlTimeResponse{
		Time: db.GetAutomatedBookShelfCrawlTime(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func setSendAlertWhenBookNoLongerAvailable(w http.ResponseWriter, r *http.Request) {
	enabled := r.URL.Query().Get("enabled")
	enabledBool := strToBool(enabled)
	if enabledBool != true && enabledBool != false {
		errorMsg := fmt.Sprintf("Invalid state '%s' given", enabled)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}

	db.SetSendAlertWhenBookNoLongerAvailable(enabledBool)
	w.WriteHeader(http.StatusOK)
}

func getSendAlertWhenBookNoLongerAvailable(w http.ResponseWriter, r *http.Request) {
	res := dtos.SendAlertWhenBookIsNoLongerAvailableResponse{
		Enabled: db.GetSendAlertWhenBookNoLongerAvailable(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func setSendAlertOnlyWhenFreeShippingKicksIn(w http.ResponseWriter, r *http.Request) {
	enabled := r.URL.Query().Get("enabled")
	enabledBool := strToBool(enabled)
	if enabledBool != true && enabledBool != false {
		errorMsg := fmt.Sprintf("Invalid state '%s' given", enabled)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetSendAlertOnlyWhenFreeShippingKicksIn(enabledBool)
	w.WriteHeader(http.StatusOK)
}

func getSendAlertOnlyWhenFreeShippingKicksIn(w http.ResponseWriter, r *http.Request) {
	res := dtos.SendAlertOnlyWhenBookFreeShippingKicksInResponse{
		Enabled: db.GetSendAlertOnlyWhenFreeShippingKicksIn(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
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
		if isStaticContent := strings.HasPrefix(r.URL.Path, "/static/"); !isStaticContent {
			logger.Sugar().Infow(fmt.Sprintf("Served request to %v", r.URL.Path),
				zap.String("requestInfo", fmt.Sprintf("%+v", r)))
		}
		next.ServeHTTP(w, r)
	})
}
