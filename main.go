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

	// engine.Worker("https://www.goodreads.com/review/list/1753152-sharon?shelf=fantasy")
	// allBooks := goodreads.GetBooksFromShelf("https://www.goodreads.com/review/list/1753152-sharon?shelf=fantasy")
	// allBooks[0] = dtos.BasicGoodReadsBook{
	// 	Title:      "The Return of the King",
	// 	Author:     "Tolkien, J.R.R.",
	// 	SeriesText: "(The Lord of the Rings, #1)",
	// }
	// searchResults := thebookshop.SearchForBooks(allBooks[:3])
	// potentialMatches := search.SearchAll(searchResults)

	// for key, potentialMatchList := range potentialMatches {
	// 	fmt.Printf("%s: ", key)
	// 	for i, potentialMatch := range potentialMatchList.SearchResultBooks {
	// 		fmt.Printf("%d - %+v", i, potentialMatch)
	// 	}
	// 	fmt.Printf("\n")
	// }

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
