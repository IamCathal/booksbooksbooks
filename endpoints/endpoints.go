package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
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
	r.HandleFunc("/status", status).Methods("POST")

	// Static content pages
	r.HandleFunc("/", index).Methods("GET")
	r.HandleFunc("/settings", settings).Methods("GET")
	r.HandleFunc("/available", available).Methods("GET")
	r.HandleFunc("/series", series).Methods("GET")

	r.HandleFunc("/getrecentcrawlbreadcrumbs", getRecentCrawlBreadcrumbs).Methods("GET")
	r.HandleFunc("/getavailablebooks", getAvailableBooks).Methods("GET")
	r.HandleFunc("/ignorebook", ignoreBook).Methods("POST")
	r.HandleFunc("/unignorebook", unignoreBook).Methods("POST")
	r.HandleFunc("/resetavailablebooks", resetAvailableBooks).Methods("POST")
	r.HandleFunc("/purgeignoredauthorsfromavailablebooks", purgeIgnoredAuthorsFromAvailableBooks).Methods("POST")
	r.HandleFunc("/getseriescrawl", getSeriesCrawl).Methods("GET")
	r.HandleFunc("/getrecentcrawlreports", getRecentCrawlReports).Methods("GET")
	r.Use(logMiddleware)

	setupWebsocketsRouter(r)
	setupSettingsRouter(r)

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
		// if isStaticContent := strings.HasPrefix(r.URL.Path, "/static/"); !isStaticContent {
		// 	logger.Sugar().Infow(fmt.Sprintf("Served request to %v", r.URL.Path),
		// 		zap.String("requestInfo", fmt.Sprintf("%+v", r)))
		// }
		next.ServeHTTP(w, r)
	})
}
