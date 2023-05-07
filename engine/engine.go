package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
	"github.com/iamcathal/booksbooksbooks/util"
	"go.uber.org/zap"
)

var (
	logger                  *zap.Logger
	FREE_SHIPPING_THRESHOLD float64 = 20.00
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func AutomatedCheckEngine() {
	for {
		currTime := controller.Cnt.GetFormattedTime()
		if currTime == db.GetAutomatedBookShelfCrawlTime() {
			logger.Info("Beginning automated crawl")
			allShelvesAreValid := true

			for _, shelfToCrawl := range db.GetShelvesToCrawl() {
				if !goodreads.CheckIsShelfURL(shelfToCrawl.ShelfURL) {
					logger.Sugar().Infof("shelfURL '%s' isn't valid", shelfToCrawl.ShelfURL)
					allShelvesAreValid = false
				}
			}

			if !allShelvesAreValid {
				logger.Info("Failed to start automated crawl because one or more shelfURLs are not valid")
			} else {
				go automatedCheck()
			}
		}
		controller.Cnt.Sleep(60 * time.Second)
	}
}

func automatedCheck() {
	stubStatsChan := make(chan int, 200)
	stubBooksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	stubSearchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

	crawlReport := dtos.AutomatedCrawlReport{
		TimeStarted: time.Now().Unix(),
	}

	booksThatWereAvailableLastTime := db.GetAvailableBooks()
	checkAvailabilityOfExistingAvailableBooksList(booksThatWereAvailableLastTime)

	booksFromShelf := goodreads.GetBooksFromShelves(db.GetShelfURLsFromShelvesToCrawl(), stubStatsChan, stubBooksFoundFromGoodReadsChan)
	logger.Sugar().Infof("Checking now for the books currently listed in shelves: %+v", db.GetShelfCrawlKeysFromShelvesToCrawl())

	// TODO move this to Generic Worke call
	// if db.GetSeriesCrawlInAutomatedCrawl() {
	// 	// extract series books and add to booksFromShelf
	// 	booksThatAreInASeries := extractGoodreadsBooksThatAreInSeries(booksFromShelf)

	// 	knownSeriesToTheirLinks := make(map[string]bool)
	// 	seriesLinks := []string{}

	// 	for _, bookInASeries := range booksThatAreInASeries {
	// 		baseSeriesTitle := goodreads.FilterSeriesTitleFromSeriesText(bookInASeries.SeriesText)
	// 		fmt.Printf("%s -> %s\n", bookInASeries.SeriesText, baseSeriesTitle)
	// 		if _, exists := knownSeriesToTheirLinks[baseSeriesTitle]; !exists {
	// 			fmt.Printf("Series %s is new, adding to the map\n", baseSeriesTitle)
	// 			knownSeriesToTheirLinks[baseSeriesTitle] = true
	// 			seriesLinks = append(seriesLinks, goodreads.GetSeriesLink(bookInASeries))
	// 		} else {
	// 			fmt.Printf("Series %s is now, not adding to the map\n", baseSeriesTitle)
	// 		}

	// 		fmt.Printf("%+v\n", knownSeriesToTheirLinks)
	// 	}

	// 	fmt.Printf("In the end here are the %d series: %+v\n", len(seriesLinks), seriesLinks)
	// 	return
	// }

	db.SetTotalBooksInAutomatedBookShelfCheck(len(booksFromShelf))
	logger.Sugar().Infof("%d books were found from shelves shelf: %+v", len(booksFromShelf), db.GetShelfCrawlKeysFromShelvesToCrawl())

	searchResults := []dtos.EnchancedSearchResult{}
	for _, book := range booksFromShelf {
		searchResults = append(searchResults, thebookshop.SearchForBook(book, stubSearchResultsFromTheBookshopChan))
		crawlReport.BooksSearched++
	}

	currentlyAvailableBooksFromShelf := goodreads.GetAvailableBooksFromSearchResult(searchResults)
	crawlReport.MatchesFound += len(currentlyAvailableBooksFromShelf)
	logger.Sugar().Infof("%d search queries were made with %d title/author matches found",
		len(searchResults), len(currentlyAvailableBooksFromShelf))

	newBooksThatNeedNotification := []dtos.AvailableBook{}
	for _, availableBook := range currentlyAvailableBooksFromShelf {
		if bookIsNew := availableBookIsNew(availableBook, booksThatWereAvailableLastTime); bookIsNew {
			newBooksThatNeedNotification = append(newBooksThatNeedNotification, availableBook)
		}
	}
	logger.Sugar().Infof("%d new books were found in this search: %v", len(newBooksThatNeedNotification),
		getConciseBookInfoFromAvailableBooks(newBooksThatNeedNotification))

	for _, newBook := range newBooksThatNeedNotification {

		logger.Sugar().Infof("Available book %s - %s was found through %d",
			newBook.BookPurchaseInfo.Author, newBook.BookPurchaseInfo.Title, newBook.BookFoundFrom)

		if authorIsIgnored := db.IsIgnoredAuthor(newBook.BookPurchaseInfo.Author); !authorIsIgnored {

			logger.Sugar().Infof("Author %s is not ignored, adding their book %s to the available book list and sending a webhook notification",
				newBook.BookPurchaseInfo.Author, newBook.BookPurchaseInfo.Title)
			db.AddAvailableBook(newBook)
			crawlReport.NewBooksFound++

			util.SendNewBookIsAvailableNotification(newBook.BookPurchaseInfo, true)
		} else {
			logger.Sugar().Infof("Author %s is ignored, their book %s will not be added to the available book list",
				newBook.BookPurchaseInfo.Author, newBook.BookPurchaseInfo.Title)
		}
	}

	logger.Info("Completed automated crawl")

	crawlReport.TimeCompleted = time.Now().Unix()
	db.AddNewRecentCrawlReport(crawlReport)

	// TODO
	// Add ws for available books
	// Add timer countdown for next automated crawl
	// Add live status for automated crawl

	sendFreeShippingWebhookIfFreeShippingEligible()
	close(stubBooksFoundFromGoodReadsChan)
	close(stubSearchResultsFromTheBookshopChan)
}

func GenericWorker(shelfURLs []string, ws *websocket.Conn) {
	previouslyKnownAvailableBooksMap := db.GetAvailableBooksMap()

	var lock sync.Mutex
	wData := dtos.WorkerInteralData{
		SearchedGoodReadsBooks: make(map[string]bool),
		SearchedSeries:         make(map[string]bool),
		Lock:                   &lock,
	}

	shelfStatsChan := make(chan int, len(shelfURLs))
	bookFoundOnGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 600)
	thebookshopSearchResultsChan := make(chan dtos.EnchancedSearchResult, 600)

	logger.Sugar().Infof("Starting crawl of %d shelves: %+v", len(shelfURLs), shelfURLs)
	goodreads.GetBooksFromShelves(shelfURLs, shelfStatsChan, bookFoundOnGoodReadsChan)

	crawlingStats := dtos.CrawlStats{
		TotalBooks:     -1,
		BooksCrawled:   0,
		BooksSearched:  0,
		BookMatchFound: 0,
	}
	var seriesLookupWorkerWg sync.WaitGroup

	for {
		if allBooksFoundInCrawl(crawlingStats) {
			logger.Sugar().Infof("All %d books found during crawl", crawlingStats.BooksCrawled)

			logger.Sugar().Infof("Waiting for all series look up worker go routines to finish")
			if waitButTimeoutAfterDuration(&seriesLookupWorkerWg, time.Duration(250*time.Millisecond)) {
				logger.Info("Some series lookup workers are still running...")
			} else {
				logger.Info("All series look up workers and books crawled, exiting")
				break
			}
		}

		select {
		case totalBooksInShelf, stillOpen := <-shelfStatsChan:
			if !stillOpen {
				shelfStatsChan = nil
			} else {
				if crawlingStats.TotalBooks == -1 {
					crawlingStats.TotalBooks = 0
				}
				crawlingStats.TotalBooks += totalBooksInShelf
				writeTotalBooksInShelfWsMessage(crawlingStats, ws)
			}

		case bookFoundOnGoodReads := <-bookFoundOnGoodReadsChan:
			wData.Lock.Lock()
			if _, exists := wData.SearchedGoodReadsBooks[bookFoundOnGoodReads.Link]; !exists {
				crawlingStats.BooksCrawled++
				wData.SearchedGoodReadsBooks[bookFoundOnGoodReads.Link] = true
				fmt.Printf("*** this book IS new %s [%s]s\n", bookFoundOnGoodReads.Title, bookFoundOnGoodReads.Link)
			} else {
				fmt.Printf("[][] this book isnt new %s [%s]\n", bookFoundOnGoodReads.Title, bookFoundOnGoodReads.Link)
				fmt.Printf("\n\n%+v\n\n", wData.SearchedGoodReadsBooks)
			}
			wData.Lock.Unlock()

			logger.Sugar().Infof("[booksCrawled: %d][booksFound: %d] Retrieved GoodReads book '%s' by '%s'",
				crawlingStats.BooksCrawled, crawlingStats.BooksSearched,
				bookFoundOnGoodReads.Title, bookFoundOnGoodReads.Author)

			go writeBookFromShelfWsMessage(bookFoundOnGoodReads, crawlingStats, ws)
			go thebookshop.SearchForBook(bookFoundOnGoodReads, thebookshopSearchResultsChan)

			if shouldSeriesSearchThisBook(bookFoundOnGoodReads, &wData) {
				bookFoundOnGoodReads.IsFromSeriesSearch = true
				seriesLookupWorkerWg.Add(1)
				logger.Sugar().Infof("Starting series worker for %s by %s", bookFoundOnGoodReads.Title, bookFoundOnGoodReads.Author)
				go SeriesLookupWorkerFunc(bookFoundOnGoodReads, bookFoundOnGoodReadsChan, &seriesLookupWorkerWg, &crawlingStats, &wData)
			}

		case searchResultFromTheBookshop := <-thebookshopSearchResultsChan:
			crawlingStats.BooksSearched++
			searchResultFiltered := filterIgnoredAuthors(searchResultFromTheBookshop)
			addAuthorsToKnownAuthors(searchResultFiltered.TitleMatches)

			allTitleAndAuthorSearchResults := searchResultFiltered.TitleMatches
			if db.AddOtherAuthorBooksIfFound() {
				allTitleAndAuthorSearchResults = append(allTitleAndAuthorSearchResults, searchResultFiltered.AuthorMatches...)
			}

			newBooksFound := getNewBooksFromSearchResult(allTitleAndAuthorSearchResults, previouslyKnownAvailableBooksMap)
			if len(newBooksFound) > 0 {
				logger.Sugar().Infof("%d new books out of %d search results were found for %s by %s", len(newBooksFound),
					len(allTitleAndAuthorSearchResults), searchResultFiltered.SearchBook.Title, searchResultFiltered.SearchBook.Author)
			}

			for _, newBook := range newBooksFound {
				crawlingStats.BookMatchFound++
				previouslyKnownAvailableBooksMap[newBook.Link] = true

				go util.SendNewBookIsAvailableNotification(newBook, true)
				writeNewAvailableBookWsMsg(newBook, crawlingStats, ws)
				db.AddAvailableBook(dtos.AvailableBook{
					BookInfo:         searchResultFromTheBookshop.SearchBook,
					BookPurchaseInfo: newBook,
					// BookFoundFrom: ,
				})
			}
			writeSearchResultReturnedMsg(searchResultFromTheBookshop, crawlingStats, ws)

		default:
			continue
		}

	}

	close(shelfStatsChan)
	close(bookFoundOnGoodReadsChan)
	close(thebookshopSearchResultsChan)

	db.SetTotalBooksInAutomatedBookShelfCheck(crawlingStats.TotalBooks)
	// addCrawlBreadcrumbsForShelves(shelves)

	logger.Sugar().Infof("Completed crawl of %d shelves, %d GoodReads books which had %d searches and %d new books",
		len(shelfURLs), crawlingStats.BooksCrawled, crawlingStats.BooksSearched, crawlingStats.BookMatchFound)
}

func Worker(shelfURL string, ws *websocket.Conn) {
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfURL '%s' given", shelfURL)
		writeErrorMsg(errorMsg, ws)
		return
	}

	// db.AddNewCrawlBreadcrumb(shelfURL)
	previouslyKnownAvailableBooksMap := db.GetAvailableBooksMap()
	previouslyKnownAvailableBooks := db.GetAvailableBooks()

	shelfStatsChan := make(chan int, 1)
	booksFromShelfChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchResultBooksChan := make(chan dtos.EnchancedSearchResult, 200)

	logger.Sugar().Infof("Retrieving books from shelf: %s", shelfURL)
	goodreads.GetBooksFromShelf(shelfURL, shelfStatsChan, booksFromShelfChan)

	booksFound := 0
	searchResultsReturned := 0
	newBooksFound := 0
	totalBooksFromGoodReads := -1

	currCrawlStats := dtos.CrawlStats{
		TotalBooks:    totalBooksFromGoodReads,
		BooksCrawled:  booksFound,
		BooksSearched: searchResultsReturned,
	}

	for {
		if allBooksFoundInCrawl(currCrawlStats) {
			break
		}

		select {
		case totalBooks, open := <-shelfStatsChan:
			if !open {
				shelfStatsChan = nil
			} else {
				currCrawlStats.TotalBooks = totalBooks
				writeTotalBooksInShelfWsMessage(currCrawlStats, ws)
			}

		case bookFromGoodReads := <-booksFromShelfChan:
			currCrawlStats.BooksCrawled++
			logger.Sugar().Infof("[booksFound: %d][booksCrawled: %d] Found a GoodReads book: %+v by %v",
				len(booksFromShelfChan), currCrawlStats.BooksCrawled,
				bookFromGoodReads.Title, bookFromGoodReads.Author)
			writeBookFromShelfWsMessage(bookFromGoodReads, currCrawlStats, ws)
			go thebookshop.SearchForBook(bookFromGoodReads, searchResultBooksChan)

		case searchResultFromTheBookshop := <-searchResultBooksChan:
			currCrawlStats.BooksSearched++

			searchResultsFiltered := filterIgnoredAuthors(searchResultFromTheBookshop)

			for _, titleMatch := range searchResultsFiltered.TitleMatches {
				db.AddAuthorToKnownAuthors(titleMatch.Author)
				if bookIsNew := wasNotPreviouslyAvailable(titleMatch, previouslyKnownAvailableBooksMap); bookIsNew {
					newBooksFound++
					currCrawlStats.BookMatchFound++
					previouslyKnownAvailableBooksMap[titleMatch.Link] = true

					logger.Sugar().Infof("Found a title match book that's for sale: %s by %s for %s at %s",
						searchResultsFiltered.SearchBook.Title, searchResultsFiltered.SearchBook.Author,
						titleMatch.Price, titleMatch.Link)

					util.SendNewBookIsAvailableNotification(titleMatch, true)

					writeNewAvailableBookWsMsg(titleMatch, currCrawlStats, ws)
					db.AddAvailableBook(dtos.AvailableBook{BookInfo: searchResultsFiltered.SearchBook, BookPurchaseInfo: titleMatch, BookFoundFrom: dtos.TITLE_MATCH})

				}
			}

			if db.AddOtherAuthorBooksIfFound() {
				for _, authorMatch := range searchResultFromTheBookshop.AuthorMatches {
					db.AddAuthorToKnownAuthors(authorMatch.Author)
					if bookIsNew := wasNotPreviouslyAvailable(authorMatch, previouslyKnownAvailableBooksMap); bookIsNew {
						newBooksFound++
						currCrawlStats.BookMatchFound++
						previouslyKnownAvailableBooksMap[authorMatch.Link] = true

						logger.Sugar().Infof("Found an author match book that's for sale: %s by %s for %s at %s",
							authorMatch.Title, authorMatch.Author, authorMatch.Price, authorMatch.Link)

						util.SendNewBookIsAvailableNotification(authorMatch, true)

						found, goodReadsListingForAuthorMatch := goodreads.SearchGoodreads(authorMatch)
						if !found {
							logger.Sugar().Warnf("Couldn't find a goodreads listing for thebookshop author match book: %+v", authorMatch)
							db.AddAvailableBook(dtos.AvailableBook{BookPurchaseInfo: authorMatch, BookFoundFrom: dtos.AUTHOR_MATCH})
							writeNewAvailableBookWsMsg(authorMatch, currCrawlStats, ws)
						} else {
							logger.Sugar().Infof("Found author match book on goodreads: %+v", goodReadsListingForAuthorMatch)

							updatedAvailableBook := dtos.AvailableBook{
								BookInfo:         goodReadsListingForAuthorMatch,
								BookPurchaseInfo: authorMatch,
								BookFoundFrom:    dtos.AUTHOR_MATCH,
								Ignore:           false,
							}
							writeNewAvailableBookWsMsg(authorMatch, currCrawlStats, ws)
							db.AddAvailableBook(updatedAvailableBook)
						}
					}
				}
			}

			writeSearchResultReturnedMsg(searchResultFromTheBookshop, currCrawlStats, ws)
		}
	}

	if notifyOnNoLongerAvailable := db.GetSendAlertWhenBookNoLongerAvailable(); notifyOnNoLongerAvailable {
		currAvailableBooks := db.GetAvailableBooks()
		if len(previouslyKnownAvailableBooks) > len(currAvailableBooks) {
			notifyAboutBooksThatAreNoLongerAvailable(previouslyKnownAvailableBooks)
		}
	}

	db.SetTotalBooksInAutomatedBookShelfCheck(currCrawlStats.TotalBooks)
	db.AddNewCrawlBreadcrumb(shelfURL, currCrawlStats.TotalBooks)

	logger.Sugar().Infof("Finished. Crawled %d books from GoodReads and made %d searches to TheBookshop.ie which had %d new books",
		currCrawlStats.BooksCrawled, currCrawlStats.BooksSearched, newBooksFound)
	close(booksFromShelfChan)
	close(searchResultBooksChan)
}

func SeriesLookupWorker(ws *websocket.Conn) []dtos.Series {
	statsChan := make(chan int, 1)
	seriesDetailsChan := make(chan dtos.Series, 200)
	initialShelfLookupChan := make(chan dtos.BasicGoodReadsBook, 200)
	lookUpBooksOnTheBookshopChan := make(chan dtos.EnchancedSearchResult, 400)

	booksFromShelf := goodreads.GetBooksFromShelves(db.GetShelfURLsFromShelvesToCrawl(), statsChan, initialShelfLookupChan)
	// close(initialShelfLookupChan)
	ownedBooksThatAreInASeries := extractGoodreadsBooksThatAreInSeries(booksFromShelf)

	fmt.Printf("%d books are in a series %+v\n", len(ownedBooksThatAreInASeries), ownedBooksThatAreInASeries)

	previouslyKnownAvailableBooksMap := db.GetAvailableBooksMap()
	knownSeriesToTheirLinks := make(map[string]bool)
	seriesLinks := []string{}

	for _, bookInASeries := range ownedBooksThatAreInASeries {
		baseSeriesTitle := goodreads.FilterSeriesTitleFromSeriesText(bookInASeries.SeriesText)
		fmt.Printf("%s -> %s\n", bookInASeries.SeriesText, baseSeriesTitle)
		if _, exists := knownSeriesToTheirLinks[baseSeriesTitle]; !exists {
			fmt.Printf("Series %s is new, adding to the map\n", baseSeriesTitle)
			knownSeriesToTheirLinks[baseSeriesTitle] = true
			seriesLinks = append(seriesLinks, goodreads.GetSeriesLink(bookInASeries))
		} else {
			fmt.Printf("Series %s is now, not adding to the map\n", baseSeriesTitle)
		}

		fmt.Printf("%+v\n", knownSeriesToTheirLinks)
	}

	fmt.Printf("Series are: %+v\n", seriesLinks)

	seriesCrawlStats := dtos.SeriesCrawlStats{
		BooksInShelf:               len(booksFromShelf),
		SeriesCount:                len(seriesLinks),
		TotalBooksInSeries:         -1,
		BooksSearchedOnTheBookshop: 0,
		SeriesLookedUp:             0,
		BookMatchesFound:           0,
	}
	writeOverallSeriesCrawlStatsMessage(seriesCrawlStats, ws)
	logger.Sugar().Infof("Found %d series (%+v) in shelf: %s\n", len(seriesLinks), seriesLinks, db.GetOwnedBooksShelfURL())

	for _, seriesLink := range seriesLinks {
		logger.Sugar().Infof("Getting series details from series link: %s", seriesLink)
		go goodreads.GetSeriesDetailsFromLink(seriesLink, seriesDetailsChan)
	}

	theBookshopMatchesFound := make(map[string]dtos.TheBookshopBook)
	shelfSeriesDetails := []dtos.Series{}

	for {
		if allBooksFoundInSeriesCrawl(seriesCrawlStats) {
			logger.Sugar().Infof("Finished series crawl: %+v", seriesCrawlStats)
			break
		}

		select {
		case seriesInfo := <-seriesDetailsChan:
			seriesCrawlStats.SeriesLookedUp++
			if seriesCrawlStats.TotalBooksInSeries == -1 {
				seriesCrawlStats.TotalBooksInSeries = 0
			}

			if db.GetOnlyEnglishBooks() {
				bookCountBeforeFiltering := len(seriesInfo.Books)
				seriesInfo.Books = filterOutNonEnglishSeriesBooks(seriesInfo.Books)
				logger.Sugar().Infof("%d non-english books were filtered out", bookCountBeforeFiltering-len(ownedBooksThatAreInASeries))
			}

			seriesCrawlStats.TotalBooksInSeries += len(seriesInfo.Books)
			writeNewSeriesFoundMessage(seriesInfo, seriesCrawlStats, ws)

			shelfSeriesDetails = append(shelfSeriesDetails, seriesInfo)
			logger.Sugar().Infof("[SeriesLookedUp: %d][BooksSearchedOnTheBookshop: %d/%d] Found series: %s which has %d books in total",
				seriesCrawlStats.SeriesLookedUp, seriesCrawlStats.BooksSearchedOnTheBookshop, seriesCrawlStats.TotalBooksInSeries, seriesInfo.Title, len(seriesInfo.Books))

			for _, bookInSeries := range seriesInfo.Books {
				go thebookshop.SearchForBook(bookInSeries.BookInfo, lookUpBooksOnTheBookshopChan)
			}

		case theBookshopSearchResult := <-lookUpBooksOnTheBookshopChan:
			seriesCrawlStats.BooksSearchedOnTheBookshop++
			logger.Sugar().Infof("[SeriesLookedUp: %d][BooksSearchedOnTheBookshop: %d/%d] Series lookup search for %s - %s had %d matches",
				seriesCrawlStats.SeriesLookedUp, seriesCrawlStats.BooksSearchedOnTheBookshop, seriesCrawlStats.TotalBooksInSeries,
				theBookshopSearchResult.SearchBook.Author, theBookshopSearchResult.SearchBook.Title,
				len(theBookshopSearchResult.TitleMatches))

			if len(theBookshopSearchResult.TitleMatches) > 0 {
				if bookIsNew := wasNotPreviouslyAvailable(theBookshopSearchResult.TitleMatches[0], previouslyKnownAvailableBooksMap); bookIsNew {
					previouslyKnownAvailableBooksMap[theBookshopSearchResult.TitleMatches[0].Link] = true
					db.AddAvailableBook(dtos.AvailableBook{BookInfo: theBookshopSearchResult.SearchBook, BookPurchaseInfo: theBookshopSearchResult.TitleMatches[0], BookFoundFrom: dtos.SERIES_MATCH})
				}
				seriesCrawlStats.BookMatchesFound++
				theBookshopMatchesFound[theBookshopSearchResult.SearchBook.Link] = theBookshopSearchResult.TitleMatches[0]
				writeSearchResultReturnedMessage(theBookshopSearchResult.SearchBook, theBookshopSearchResult.TitleMatches[0], seriesCrawlStats, ws)
			} else {
				writeSearchResultReturnedMessage(theBookshopSearchResult.SearchBook, dtos.TheBookshopBook{}, seriesCrawlStats, ws)
			}

			// TODO add author matches here if its not already done
			for _, authorMatch := range theBookshopSearchResult.AuthorMatches {
				if bookIsNew := wasNotPreviouslyAvailable(authorMatch, previouslyKnownAvailableBooksMap); bookIsNew {
					previouslyKnownAvailableBooksMap[authorMatch.Link] = true
					db.AddAvailableBook(dtos.AvailableBook{BookInfo: theBookshopSearchResult.SearchBook, BookPurchaseInfo: authorMatch, BookFoundFrom: dtos.SERIES_MATCH})
				}
				seriesCrawlStats.BookMatchesFound++
				theBookshopMatchesFound[theBookshopSearchResult.SearchBook.Link] = authorMatch
				if db.AddOtherAuthorBooksIfFound() {
					writeSearchResultReturnedMessage(theBookshopSearchResult.SearchBook, authorMatch, seriesCrawlStats, ws)
				} else {
					writeSearchResultReturnedMessage(theBookshopSearchResult.SearchBook, dtos.TheBookshopBook{}, seriesCrawlStats, ws)
				}
			}

		}
	}
	close(lookUpBooksOnTheBookshopChan)
	close(seriesDetailsChan)
	close(initialShelfLookupChan)

	logger.Sugar().Infof("Found %d matches out of %d searches for series crawl lookup of shelf: %s\n",
		seriesCrawlStats.BookMatchesFound, seriesCrawlStats.BooksSearchedOnTheBookshop, db.GetOwnedBooksShelfURL())

	updatedShelfSeriesDetailsWithMatches := shelfSeriesDetails

	for bookLink, match := range theBookshopMatchesFound {
		logger.Sugar().Infof("Inserting match %s -> %+v", bookLink, match)

		for k, bookSeries := range shelfSeriesDetails {
			for i := 0; i < len(bookSeries.Books); i++ {
				currBook := bookSeries.Books[i]
				if _, exists := theBookshopMatchesFound[currBook.BookInfo.Link]; exists {
					updatedShelfSeriesDetailsWithMatches[k].Books[i].TheBookshopMatch = theBookshopMatchesFound[currBook.BookInfo.Link]
				}
			}
		}
	}

	db.SetSeriesCrawlBooks(updatedShelfSeriesDetailsWithMatches)
	logger.Info("Done series crawl :)")
	return updatedShelfSeriesDetailsWithMatches
}

func SeriesLookupWorkerFunc(book dtos.BasicGoodReadsBook, bookFoundOnGoodReadsChan chan<- dtos.BasicGoodReadsBook, waitG *sync.WaitGroup, crawlStats *dtos.CrawlStats, wData *dtos.WorkerInteralData) {
	defer waitG.Done()
	baseSeriesTitle := goodreads.FilterSeriesTitleFromSeriesText(book.SeriesText)
	logger.Sugar().Infof("Looking up series info for '%s' by '%s' in series: '%s'", book.Title, book.Author, baseSeriesTitle)

	seriesDetails := goodreads.GetSeriesDetails(book)
	logger.Sugar().Infof("Series '%s' had %d books", baseSeriesTitle, len(seriesDetails.Books))

	wData.Lock.Lock()
	defer wData.Lock.Unlock()
	for _, book := range seriesDetails.Books {
		bookLink := book.BookInfo.Link
		if _, exists := wData.SearchedGoodReadsBooks[bookLink]; !exists {
			crawlStats.TotalBooks++
			bookFoundOnGoodReadsChan <- book.BookInfo
		}
	}
}

func allBooksFoundInSeriesCrawl(crawlStats dtos.SeriesCrawlStats) bool {
	return crawlStats.BooksSearchedOnTheBookshop == crawlStats.TotalBooksInSeries &&
		(crawlStats.SeriesLookedUp == crawlStats.SeriesCount) &&
		crawlStats.TotalBooksInSeries != -1
}

func allBooksFoundInCrawl(crawlStats dtos.CrawlStats) bool {
	// fmt.Printf("\n\nbooks crawled: %d\ntotalbooks %d\nbooks searched: %d\n\n", crawlStats.BooksCrawled, crawlStats.TotalBooks, crawlStats.BooksSearched)
	if ((crawlStats.BooksCrawled == crawlStats.TotalBooks) &&
		(crawlStats.BooksSearched == crawlStats.TotalBooks)) && crawlStats.TotalBooks != -1 {
		return true
	}
	return false
}

func availableBookIsNew(newBook dtos.AvailableBook, oldList []dtos.AvailableBook) bool {
	for _, book := range oldList {
		if book.BookInfo.Title == newBook.BookInfo.Title {
			return false
		}
	}
	return true
}

func shouldSeriesSearchThisBook(bookFoundOnGoodReads dtos.BasicGoodReadsBook, wData *dtos.WorkerInteralData) bool {
	if bookFoundOnGoodReads.IsFromSeriesSearch {
		return false
	}
	return bookFoundOnGoodReads.SeriesText != "" &&
		db.SearchOtherSeriesBooksLookup() &&
		isNewSeries(bookFoundOnGoodReads.SeriesText, wData)
}

func isNewSeries(seriesText string, wData *dtos.WorkerInteralData) bool {
	wData.Lock.Lock()
	defer wData.Lock.Unlock()

	baseSeriesText := goodreads.FilterSeriesTitleFromSeriesText(seriesText)
	if _, exists := wData.SearchedSeries[baseSeriesText]; !exists {
		wData.SearchedSeries[baseSeriesText] = true
		fmt.Printf("%s IS a new series\n", baseSeriesText)
		return true
	}
	fmt.Printf("%s is not a new series\n", baseSeriesText)
	return false
}
