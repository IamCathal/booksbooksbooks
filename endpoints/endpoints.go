package endpoints

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	appConfig dtos.AppConfig
)

func InitConfig(conf dtos.AppConfig) {
	appConfig = conf
}

func SetupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/status", status).Methods("POST")
	r.Use(logMiddleware)
	return r
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
		fmt.Printf("%v %+v\n", time.Now().Format(time.RFC3339), r)
		next.ServeHTTP(w, r)
	})
}
