package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
)

func setupSettingsRouter(mainRouter *mux.Router) *mux.Router {
	settingsRouter := mainRouter.PathPrefix("/settings").Subrouter()
	settingsRouter.HandleFunc("/getpreviewforshelf", getPreviewForShelf).Methods("GET")
	settingsRouter.HandleFunc("/getshelvestocrawl", getPreviewsForShelves).Methods("GET")
	settingsRouter.HandleFunc("/addshelftocrawl", addShelfToCrawl).Methods("POST")
	settingsRouter.HandleFunc("/removeshelftocrawl", removeShelfToCrawl).Methods("POST")

	settingsRouter.HandleFunc("/setdiscordmessageformat", setDiscordMessageFormat).Methods("POST")
	settingsRouter.HandleFunc("/getdiscordmessageformat", getDiscordMessageFormat).Methods("GET")

	settingsRouter.HandleFunc("/setautomatedcrawltime", setAutomatedCrawlTime).Methods("POST")
	settingsRouter.HandleFunc("/getautomatedcrawltime", getAutomatedCrawlTime).Methods("GET")
	settingsRouter.HandleFunc("/disableautomatedcrawltime", disableAutomatedCrawlTime).Methods("POST")

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

	return settingsRouter
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

	res := dtos.GetPreviewForShelfResponse{
		ShelfToCrawl: goodreads.GenerateShelfToCrawlEntryAndSave(shelfURL),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func getPreviewsForShelves(w http.ResponseWriter, r *http.Request) {
	res := dtos.GetPreviewsForShelvesResponse{
		ShelvesToCrawl: db.GetShelvesToCrawl(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func addShelfToCrawl(w http.ResponseWriter, r *http.Request) {
	shelfURL := r.URL.Query().Get("shelfurl")
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfurl '%s' given", shelfURL)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	goodreads.GenerateShelfToCrawlEntryAndSave(shelfURL)

	w.WriteHeader(http.StatusOK)
}

func removeShelfToCrawl(w http.ResponseWriter, r *http.Request) {
	shelfURL := r.URL.Query().Get("shelfurl")
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfurl '%s' given", shelfURL)
		SendBasicInvalidResponse(w, r, errorMsg, http.StatusBadRequest)
		return
	}
	db.RemoveShelfFromShelvesToCrawl(shelfURL)

	w.WriteHeader(http.StatusOK)
}

func getRecentCrawlBreadcrumbs(w http.ResponseWriter, r *http.Request) {
	recentCrawlbreadcrumbs := db.GetRecentCrawlBreadcrumbs()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(recentCrawlbreadcrumbs)
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
