package main

import (
	"time"

	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
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

	goodreads.GetBooksFromShelf("https://www.goodreads.com/review/list/1753152-sharon?shelf=fantasy")

	// appConfig := initConfig()
	// endpoints.InitConfig(appConfig)
	// port := "2945"

	// router := endpoints.SetupRouter()

	// srv := &http.Server{
	// 	Handler:      router,
	// 	Addr:         ":" + fmt.Sprint(port),
	// 	WriteTimeout: 10 * time.Second,
	// 	ReadTimeout:  10 * time.Second,
	// }
	// fmt.Println("serving requests on :" + port)
	// log.Fatal(srv.ListenAndServe())
}
