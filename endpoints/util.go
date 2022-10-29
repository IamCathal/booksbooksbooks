package endpoints

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func DisallowFileBrowsing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = filepath.Clean(r.URL.Path)
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/static") {
			next.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	})
}

func setupWebSocket(w http.ResponseWriter, r *http.Request) *websocket.Conn {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// if _, ok := err.(websocket.HandshakeError); !ok {
		// 	return nil
		// }
		return nil
	}
	return ws
}

func SendBasicInvalidResponse(w http.ResponseWriter, req *http.Request, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	response := struct {
		Error string `json:"error"`
	}{
		msg,
	}
	json.NewEncoder(w).Encode(response)
}

func isActualEndpoint(urlPath string) bool {
	regularEndpoints := []string{
		"/",
		"/status",
		"/getrecentcrawls",
		"/ws",
		"/available",
		"/automatedcheck",
		"/getavailablebooks",
	}
	for _, endpoint := range regularEndpoints {
		if urlPath == endpoint {
			return true
		}
	}
	return false
}

func bookIsNew(newBook dtos.AvailableBook, oldList []dtos.AvailableBook) bool {
	for _, book := range oldList {
		if book.BookInfo.Title == newBook.BookInfo.Title {
			return false
		}
	}
	return true
}
