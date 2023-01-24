package engine

import (
	"fmt"
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
	logger                   *zap.Logger
	BOOKS_DISPLAYED_PER_PAGE         = 30
	FREE_SHIPPING_THRESHOLD  float64 = 20.00
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

	db.SetTotalBooksInAutomatedBookShelfCheck(len(booksFromShelf))
	logger.Sugar().Infof("%d books were found from shelves shelf: %+v\n", len(booksFromShelf), db.GetShelfCrawlKeysFromShelvesToCrawl())

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

	// Add ws for available books
	// Add timer countdown for next automated crawl
	// Add live status for automated crawl

	sendFreeShippingWebhookIfFreeShippingEligible()
	close(stubBooksFoundFromGoodReadsChan)
	close(stubSearchResultsFromTheBookshopChan)
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
		if allBooksFoundInShelfCrawl(currCrawlStats) {
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

			if addMoreAuthorBooksToAvailableBooksList := db.GetAddMoreAuthorBooksToAvailableBooksList(); addMoreAuthorBooksToAvailableBooksList {
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

	booksFromShelf := goodreads.GetBooksFromShelf(db.GetOwnedBooksShelfURL(), statsChan, initialShelfLookupChan)
	close(initialShelfLookupChan)
	ownedBooksThatAreInASeries := extractGoodreadsBooksThatAreInSeries(booksFromShelf)

	previouslyKnownAvailableBooksMap := db.GetAvailableBooksMap()
	knownSeriesToTheirLinks := make(map[string]bool)
	seriesLinks := []string{}

	for _, bookInASeries := range ownedBooksThatAreInASeries {
		baseSeriesTitle := goodreads.FilterSeriesTitleFromSeriesText(bookInASeries.SeriesText)
		if _, exists := knownSeriesToTheirLinks[baseSeriesTitle]; !exists {
			knownSeriesToTheirLinks[baseSeriesTitle] = true
			seriesLinks = append(seriesLinks, goodreads.GetSeriesLink(bookInASeries))
		}
	}

	seriesCrawlStats := dtos.SeriesCrawlStats{
		BooksInShelf:               len(booksFromShelf),
		SeriesCount:                len(seriesLinks),
		TotalBooksInSeries:         -1,
		BooksSearchedOnTheBookshop: 0,
		SeriesLookedUp:             0,
		BookMatchesFound:           0,
	}
	writeOverallSeriesCrawlStatsMessage(seriesCrawlStats, ws)
	logger.Sugar().Infof("Found %d series in shelf: %s\n", len(seriesLinks), db.GetOwnedBooksShelfURL())

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
			seriesCrawlStats.TotalBooksInSeries += len(seriesInfo.Books)
			writeNewSeriesFoundMessage(seriesInfo, seriesCrawlStats, ws)

			shelfSeriesDetails = append(shelfSeriesDetails, seriesInfo)
			logger.Sugar().Infof("[SeriesLookedUp: %d][BooksSearchedOnTheBookshop: %d/%d] Found series: %s which have %d books in total",
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
		}
	}
	close(lookUpBooksOnTheBookshopChan)
	close(seriesDetailsChan)

	logger.Sugar().Infof("Found %d matches out of %d searches for series crawl lookup of shelf: %s\n",
		seriesCrawlStats.BookMatchesFound, seriesCrawlStats.BooksSearchedOnTheBookshop, db.GetOwnedBooksShelfURL())

	updatedShelfSeriesDetailsWithMatches := shelfSeriesDetails

	for bookLink, match := range theBookshopMatchesFound {
		logger.Sugar().Infof("Inserting match %s -> %+v into series details: %+v", bookLink, match)

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
	return updatedShelfSeriesDetailsWithMatches
}

func allBooksFoundInSeriesCrawl(crawlStats dtos.SeriesCrawlStats) bool {
	return crawlStats.BooksSearchedOnTheBookshop == crawlStats.TotalBooksInSeries &&
		(crawlStats.SeriesLookedUp == crawlStats.SeriesCount) &&
		crawlStats.TotalBooksInSeries != -1
}

func allBooksFoundInShelfCrawl(crawlStats dtos.CrawlStats) bool {
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
