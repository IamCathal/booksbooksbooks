package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/endpoints"
)

var (
	ApplicationStartUpTime time.Time
)

func initConfig() dtos.AppConfig {
	return dtos.AppConfig{
		ApplicationStartUpTime: time.Now(),
	}
}

func main() {
	appConfig := initConfig()
	endpoints.InitConfig(appConfig)
	port := "2945"

	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + fmt.Sprint(port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	fmt.Println("serving requests on :" + port)
	log.Fatal(srv.ListenAndServe())
}
