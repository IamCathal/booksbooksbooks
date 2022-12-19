package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

var (
	logger             *zap.Logger
	websocketWriteLock sync.Mutex
	Cnt                CntrInterface
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func SetController(controller CntrInterface) {
	Cnt = controller
}

type Cntr struct{}

type CntrInterface interface {
	// Goodreads and TheBookshop
	GetPage(url string) *html.Node
	Get(pageURL string) []byte

	// Websocket and notifications
	WriteWsMessage(msg []byte, ws *websocket.Conn) error
	DeliverWebhook(msg dtos.DiscordMsg) error

	// Utils
	GetFormattedTime() string
	Sleep(duration time.Duration)
}

func (control Cntr) GetPage(pageURL string) *html.Node {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	if contains := strings.Contains(pageURL, "thebookshop.ie"); contains {
		setBookBookBooksHeaders(req)
	}
	if contains := strings.Contains(pageURL, "goodreads.com"); contains {
		setGoodreadsHeaders(req, pageURL)
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	doc, err := html.Parse(res.Body)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	return doc
}

func (control Cntr) Get(pageURL string) []byte {
	fmt.Printf("Get: '%s'\n", pageURL)
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageURL, nil)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	if contains := strings.Contains(pageURL, "thebookshop.ie"); contains {
		setBookBookBooksHeaders(req)
	}
	if contains := strings.Contains(pageURL, "goodreads.com"); contains {
		setGoodreadsHeaders(req, pageURL)
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	// fmt.Printf("The HTML: \n\n%s\n\n", string(body))
	return body
}

func (control Cntr) WriteWsMessage(msg []byte, ws *websocket.Conn) error {
	websocketWriteLock.Lock()
	defer websocketWriteLock.Unlock()
	return ws.WriteMessage(1, msg)
}

func (control Cntr) DeliverWebhook(msg dtos.DiscordMsg) error {
	webhookURL := db.GetDiscordWebhookURL()
	if webhookURL == "" {
		return nil
	}

	msgEmbedByte, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(msgEmbedByte))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	return err
}

func (control Cntr) GetFormattedTime() string {
	now := time.Now()

	currHour := now.Hour()
	currHourFormatted := ""
	if currHour >= 0 && currHour <= 9 {
		currHourFormatted = fmt.Sprintf("0%d", currHour)
	} else {
		currHourFormatted = fmt.Sprint(now.Hour())
	}

	currMinute := now.Minute()
	currMinuteFormatted := ""
	if currMinute >= 0 && currMinute <= 9 {
		currMinuteFormatted += fmt.Sprintf("0%d", currMinute)
	} else {
		currMinuteFormatted = fmt.Sprint(now.Minute())
	}

	return fmt.Sprintf("%s:%s", currHourFormatted, currMinuteFormatted)
}

func (control Cntr) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func setGoodreadsHeaders(req *http.Request, url string) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Host", "www.goodreads.com")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Referer", getFakeGoodreadsReferrerPage(url))
}

func setBookBookBooksHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Alt-Used", "cdn11.bigcommerce.com")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Host", "cdn11.bigcommerce.com")
	req.Header.Set("TE", "trailers")
	req.Header.Set("Referer", "https://thebookshop.ie/")
}

func getFakeGoodreadsReferrerPage(URL string) string {
	splitStringByShelfParam := strings.Split(URL, "?")
	return splitStringByShelfParam[0]
}
