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
			shelfURL := db.GetAutomatedBookShelfCheck()
			if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
				logger.Sugar().Infof("Failed to start automated crawl because shelfURL '%s' isn't valid", shelfURL)
			} else {
				go automatedCheck()
			}
		}
		controller.Cnt.Sleep(60 * time.Second)
	}
}

func automatedCheck() {
	stubStatsChan := make(chan int, 1)
	stubBooksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	stubSearchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

	booksThatWereAvailableLastTime := db.GetAvailableBooks()
	checkAvailabilityOfExistingAvailableBooksList()

	booksFromShelf := goodreads.GetBooksFromShelf(db.GetAutomatedBookShelfCheck(), stubStatsChan, stubBooksFoundFromGoodReadsChan)
	db.SetTotalBooksInAutomatedBookShelfCheck(len(booksFromShelf))

	logger.Sugar().Infof("%d books were found from GoodReads shelf: %s\n", len(booksFromShelf), db.GetAutomatedBookShelfCheck())
	close(stubBooksFoundFromGoodReadsChan)

	searchResults := []dtos.EnchancedSearchResult{}
	for _, book := range booksFromShelf {
		searchResults = append(searchResults, thebookshop.SearchForBook(book, stubSearchResultsFromTheBookshopChan))
	}
	currentlyAvailableBooksFromShelf := goodreads.GetAvailableBooksFromSearchResult(searchResults)
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
		if authorIsIgnored := db.IsIgnoredAuthor(newBook.BookPurchaseInfo.Author); !authorIsIgnored {
			db.AddAvailableBook(newBook)
			util.SendNewBookIsAvailableNotification(newBook.BookPurchaseInfo, true)
		}
	}

	sendFreeShippingWebhookIfFreeShippingEligible()
}

func Worker(shelfURL string, ws *websocket.Conn) {
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfURL '%s' given", shelfURL)
		writeErrorMsg(errorMsg, ws)
		return
	}

	db.AddNewCrawlBreadcrumb(shelfURL)
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
		if allBooksFound(currCrawlStats) {
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

	logger.Sugar().Infof("Finished. Crawled %d books from GoodReads and made %d searches to TheBookshop.ie which had %d new books",
		currCrawlStats.BooksCrawled, currCrawlStats.BooksSearched, newBooksFound)
	close(booksFromShelfChan)
	close(searchResultBooksChan)
}

func allBooksFound(crawlStats dtos.CrawlStats) bool {
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
