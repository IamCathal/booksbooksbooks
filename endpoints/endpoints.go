package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/engine"
	"github.com/iamcathal/booksbooksbooks/goodreads"
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
	r.HandleFunc("/status", status).Methods("POST")
	r.HandleFunc("/settings", settings).Methods("GET")
	r.HandleFunc("/available", available).Methods("GET")
	r.HandleFunc("/series", series).Methods("GET")
	r.HandleFunc("/getrecentcrawls", getRecentCrawls).Methods("GET")
	r.HandleFunc("/getavailablebooks", getAvailableBooks).Methods("GET")
	r.HandleFunc("/ignorebook", ignoreBook).Methods("POST")
	r.HandleFunc("/unignorebook", unignoreBook).Methods("POST")
	r.HandleFunc("/resetavailablebooks", resetAvailableBooks).Methods("POST")
	r.HandleFunc("/purgeignoredauthorsfromavailablebooks", purgeIgnoredAuthorsFromAvailableBooks).Methods("POST")
	r.HandleFunc("/getautomatedcrawlshelfstats", getAutomatedCrawlShelfStats).Methods("GET")
	r.HandleFunc("/getseriescrawl", getSeriesCrawl).Methods("GET")
	r.Use(logMiddleware)

	websocketRouter := r.PathPrefix("/ws").Subrouter()
	websocketRouter.HandleFunc("/shelfcrawl", shelfCrawl).Methods("GET")
	websocketRouter.HandleFunc("/seriescrawl", seriesCrawl).Methods("GET")

	settingsRouter := r.PathPrefix("/settings").Subrouter()
	settingsRouter.HandleFunc("/getpreviewforshelf", getPreviewForShelf).Methods("GET")
	settingsRouter.HandleFunc("/setdiscordmessageformat", setDiscordMessageFormat).Methods("POST")
	settingsRouter.HandleFunc("/getdiscordmessageformat", getDiscordMessageFormat).Methods("GET")
	settingsRouter.HandleFunc("/setautomatedcrawltime", setAutomatedCrawlTime).Methods("POST")
	settingsRouter.HandleFunc("/getautomatedcrawltime", getAutomatedCrawlTime).Methods("GET")
	settingsRouter.HandleFunc("/disableautomatedcrawltime", disableAutomatedCrawlTime).Methods("POST")
	settingsRouter.HandleFunc("/getautomatedbookshelfcheckurl", getautomatedbookshelfcheckurl).Methods("GET")
	settingsRouter.HandleFunc("/setautomatedbookshelfcheckurl", setautomatedbookshelfcheckurl).Methods("POST")
	settingsRouter.HandleFunc("/testdiscordwebhook", testDiscordWebook).Methods("GET")
	settingsRouter.HandleFunc("/setdiscordwebhook", setDiscordWebook).Methods("POST")
	settingsRouter.HandleFunc("/getdiscordwebhook", getDiscordWebook).Methods("GET")
	settingsRouter.HandleFunc("/cleardiscordwebhook", clearDiscordWebhook).Methods("POST")
	settingsRouter.HandleFunc("/setsendalertwhenbooknolongeravailable", setSendAlertWhenBookNoLongerAvailable).Methods("POST")
	settingsRouter.HandleFunc("/getsendalertwhenbooknolongeravailable", getSendAlertWhenBookNoLongerAvailable).Methods("GET")
	settingsRouter.HandleFunc("/setsendalertonlywhenfreeshippingkicksin", setSendAlertOnlyWhenFreeShippingKicksIn).Methods("POST")
	settingsRouter.HandleFunc("/getsendalertonlywhenfreeshippingkicksin", getSendAlertOnlyWhenFreeShippingKicksIn).Methods("GET")
	settingsRouter.HandleFunc("/setaddmoreauthorbookstoavailablelist", setAddMoreAuthorBooksToAvailableBooksList).Methods("POST")
	settingsRouter.HandleFunc("/getaddmoreauthorbookstoavailablelist", getAddMoreAuthorBooksToAvailableBooksList).Methods("GET")
	settingsRouter.HandleFunc("/getknownauthors", getKnownAuthors).Methods("Get")
	settingsRouter.HandleFunc("/clearknownauthors", clearKnownAuthors).Methods("POST")
	settingsRouter.HandleFunc("/toggleauthorignore", toggleAuthorIgnore).Methods("POST")
	settingsRouter.HandleFunc("/getownedbooksshelfurl", getOwnedBooksShelfURL).Methods("GET")
	settingsRouter.HandleFunc("/setownedbooksshelfurl", setOwnedBooksShelfURL).Methods("POST")
	settingsRouter.HandleFunc("/getonlyenglishbooksenabled", getOnlyEnglishBooksEnabled).Methods("GET")
	settingsRouter.HandleFunc("/setonlyenglishbooksenabled", setOnlyEnglishBooksEnabled).Methods("POST")
	settingsRouter.HandleFunc("/purgeauthormatches", purgeAuthorMatches).Methods("POST")
	settingsRouter.HandleFunc("/purgeseriesmatches", purgeSeriesMatches).Methods("POST")
	settingsRouter.HandleFunc("/getseriesinautomatedcrawl", getSeriesInAutomatedCrawl).Methods("GET")
	settingsRouter.HandleFunc("/setseriesinautomatedcrawl", setSeriesInAutomatedCrawl).Methods("POST")
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

func series(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/series.html")
}

func settings(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/settings.html")
}

func shelfCrawl(w http.ResponseWriter, r *http.Request) {
	ws := setupWebSocket(w, r)
	if ws == nil {
		SendBasicInvalidResponse(w, r, "unable to upgrade websocket", http.StatusBadRequest)
		return
	}
	engine.Worker(r.URL.Query().Get("shelfurl"), ws)
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
	_, err := url.Parse(discordWebhook)
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid webhookurl '%s' is not a valid URL", discordWebhook)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetDiscordWebhookURL(discordWebhook)
	err = controller.Cnt.DeliverWebhook(dtos.DiscordMsg{
		Username:   "BooksBooksBooks",
		Avatar_url: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
		Embed: []dtos.DiscordEmbed{
			{
				Title: "BooksBooksBooks is ready to send webhook updates to this channel",
			},
		},
	})
	if err != nil {
		errorMsg := fmt.Sprintf("Invalid webhookurl '%s' given", discordWebhook)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
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

func clearDiscordWebhook(w http.ResponseWriter, r *http.Request) {
	db.SetDiscordWebhookURL("")
	w.WriteHeader(http.StatusOK)
}

func resetAvailableBooks(w http.ResponseWriter, r *http.Request) {
	db.ResetAvailableBooks()
	w.WriteHeader(http.StatusOK)
}

func purgeIgnoredAuthorsFromAvailableBooks(w http.ResponseWriter, r *http.Request) {
	db.PurgeIgnoredAuthorsFromAvailableBooks()
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

func getSeriesCrawl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(db.GetSeriesCrawlBooks())
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

func disableAutomatedCrawlTime(w http.ResponseWriter, r *http.Request) {
	db.SetAutomatedBookShelfCrawlTime("")
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
	enabledBool, isValid := strToBool(enabled)
	if !isValid {
		errorMsg := fmt.Sprintf("Invalid state '%s' given", enabled)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}

	db.SetSendAlertWhenBookNoLongerAvailable(enabledBool)
	w.WriteHeader(http.StatusOK)
}

func getSendAlertWhenBookNoLongerAvailable(w http.ResponseWriter, r *http.Request) {
	res := dtos.BooleanSettingStatusResponse{
		Enabled: db.GetSendAlertWhenBookNoLongerAvailable(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func setSendAlertOnlyWhenFreeShippingKicksIn(w http.ResponseWriter, r *http.Request) {
	enabled := r.URL.Query().Get("enabled")
	enabledBool, isValid := strToBool(enabled)
	if !isValid {
		errorMsg := fmt.Sprintf("Invalid state '%s' given", enabled)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetSendAlertOnlyWhenFreeShippingKicksIn(enabledBool)
	w.WriteHeader(http.StatusOK)
}

func getSendAlertOnlyWhenFreeShippingKicksIn(w http.ResponseWriter, r *http.Request) {
	res := dtos.BooleanSettingStatusResponse{
		Enabled: db.GetSendAlertOnlyWhenFreeShippingKicksIn(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func setAddMoreAuthorBooksToAvailableBooksList(w http.ResponseWriter, r *http.Request) {
	enabled := r.URL.Query().Get("enabled")
	enabledBool, isValid := strToBool(enabled)
	if !isValid {
		errorMsg := fmt.Sprintf("Invalid state '%s' given", enabled)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetAddMoreAuthorBooksToAvailableBooksList(enabledBool)
	w.WriteHeader(http.StatusOK)
}

func getAddMoreAuthorBooksToAvailableBooksList(w http.ResponseWriter, r *http.Request) {
	res := dtos.BooleanSettingStatusResponse{
		Enabled: db.GetAddMoreAuthorBooksToAvailableBooksList(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func getKnownAuthors(w http.ResponseWriter, r *http.Request) {
	res := db.GetKnownAuthors()
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func clearKnownAuthors(w http.ResponseWriter, r *http.Request) {
	db.SetKnownAuthors([]dtos.KnownAuthor{})
	w.WriteHeader(http.StatusOK)
}

func toggleAuthorIgnore(w http.ResponseWriter, r *http.Request) {
	author := r.URL.Query().Get("author")
	if author == "" {
		errorMsg := fmt.Sprintf("Invalid author '%s' given", author)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.ToggleAuthorIgnore(author)
	w.WriteHeader(http.StatusOK)
}

func getOwnedBooksShelfURL(w http.ResponseWriter, r *http.Request) {
	bookShelfURL := db.GetOwnedBooksShelfURL()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(dtos.AutomatedShelfCheckURLResponse{ShelURL: bookShelfURL})
}

func setOwnedBooksShelfURL(w http.ResponseWriter, r *http.Request) {
	bookShelfURL := r.URL.Query().Get("shelfurl")
	if isValidShelfURL := goodreads.CheckIsShelfURL(bookShelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfurl '%s' given", bookShelfURL)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}

	db.SetOwnedBooksShelfURL(bookShelfURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookShelfURL)
}

func getOnlyEnglishBooksEnabled(w http.ResponseWriter, r *http.Request) {
	res := dtos.BooleanSettingStatusResponse{
		Enabled: db.GetOnlyEnglishBooks(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func setOnlyEnglishBooksEnabled(w http.ResponseWriter, r *http.Request) {
	enabled := r.URL.Query().Get("enabled")
	enabledBool, isValid := strToBool(enabled)
	if !isValid {
		errorMsg := fmt.Sprintf("Invalid state '%s' given", enabled)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetOnlyEnglishBooks(enabledBool)
	w.WriteHeader(http.StatusOK)
}

func purgeAuthorMatches(w http.ResponseWriter, r *http.Request) {
	allAvailableBooks := db.GetAvailableBooks()
	availableBooksThatAreNotAuthorMatches := []dtos.AvailableBook{}
	allAvailableBooksCount := len(allAvailableBooks)

	for _, book := range allAvailableBooks {
		if book.BookFoundFrom != dtos.AUTHOR_MATCH {
			availableBooksThatAreNotAuthorMatches = append(availableBooksThatAreNotAuthorMatches, book)
		}
	}
	db.SetAvailableBooks(availableBooksThatAreNotAuthorMatches)
	logger.Sugar().Infof("Purged %d author matches", allAvailableBooksCount-len(availableBooksThatAreNotAuthorMatches))

	w.WriteHeader(http.StatusOK)
}

func purgeSeriesMatches(w http.ResponseWriter, r *http.Request) {
	allAvailableBooks := db.GetAvailableBooks()
	availableBooksThatAreNotSeriesMatches := []dtos.AvailableBook{}
	allAvailableBooksCount := len(allAvailableBooks)

	for _, book := range allAvailableBooks {
		if book.BookFoundFrom != dtos.SERIES_MATCH {
			availableBooksThatAreNotSeriesMatches = append(availableBooksThatAreNotSeriesMatches, book)
		}
	}
	db.SetAvailableBooks(availableBooksThatAreNotSeriesMatches)
	logger.Sugar().Infof("Purged %d series matches", allAvailableBooksCount-len(availableBooksThatAreNotSeriesMatches))

	w.WriteHeader(http.StatusOK)
}

func getSeriesInAutomatedCrawl(w http.ResponseWriter, r *http.Request) {
	res := dtos.BooleanSettingStatusResponse{
		Enabled: db.GetSeriesCrawlInAutomatedCrawl(),
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func setSeriesInAutomatedCrawl(w http.ResponseWriter, r *http.Request) {
	enabled := r.URL.Query().Get("enabled")
	enabledBool, isValid := strToBool(enabled)
	if !isValid {
		errorMsg := fmt.Sprintf("Invalid state '%s' given", enabled)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.SetSeriesCrawlInAutomatedCrawl(enabledBool)
	w.WriteHeader(http.StatusOK)
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
