package engine

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/iamcathal/booksbooksbooks/goodreads"
	"github.com/iamcathal/booksbooksbooks/thebookshop"
	"github.com/iamcathal/booksbooksbooks/util"
	"go.uber.org/zap"
)

var (
	logger                   *zap.Logger
	BOOKS_DISPLAYED_PER_PAGE = 30
)

func SetLogger(newLogger *zap.Logger) {
	logger = newLogger
}

func AutomatedCheckEngine() {
	for {
		currTime := getFormattedTime()
		if currTime == db.GetAutomatedBookShelfCrawlTime() {
			logger.Info("Beginning automated crawl")
			shelfURL := db.GetAutomatedBookShelfCheck()
			if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
				logger.Sugar().Infof("Failed to start automated crawl because shelfURL '%s' isn't valid", shelfURL)
			} else {
				go automatedCheck()
			}
		}
		time.Sleep(60 * time.Second)
	}
}

func automatedCheck() {
	stubStatsChan := make(chan int, 1)
	stubBooksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	stubSearchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

	cachedBooksThatWereAvailable := db.GetAvailableBooks()
	cachedBooksThatAreStillAvailableToday := []dtos.AvailableBook{}
	booksFromShelfThatAreAvailableNow := []dtos.AvailableBook{}

	for _, book := range cachedBooksThatWereAvailable {
		searchResult := thebookshop.SearchForBook(book.BookInfo, stubSearchResultsFromTheBookshopChan)

		if len(searchResult.TitleMatches) >= 1 {
			cachedBooksThatAreStillAvailableToday = append(cachedBooksThatAreStillAvailableToday, book)
		}
	}

	logger.Sugar().Infof("%d cached books that were available from the last automated checkup: %d\n",
		len(cachedBooksThatWereAvailable), cachedBooksThatWereAvailable)
	logger.Sugar().Infof("%d Cached from from the last automated checkup that are still available now: %d\n",
		len(cachedBooksThatAreStillAvailableToday), cachedBooksThatAreStillAvailableToday)

	if alertOnNoLongerAvailableBooks := db.GetSendAlertWhenBookNoLongerAvailable(); alertOnNoLongerAvailableBooks {
		booksThatAreNowNotAvailable := util.FindBooksThatAreNowNotAvailable(cachedBooksThatWereAvailable, cachedBooksThatAreStillAvailableToday)
		for _, book := range booksThatAreNowNotAvailable {
			util.SendNewBookIsAvailableMessage(book.BookPurchaseInfo)
		}
	}

	shelfURL := db.GetAutomatedBookShelfCheck()

	booksFromShelf := goodreads.GetBooksFromShelf(shelfURL, stubStatsChan, stubBooksFoundFromGoodReadsChan)
	db.SetTotalBooksInAutomatedBookShelfCheck(len(booksFromShelf))
	logger.Sugar().Infof("%d books were found from GoodReads shelf %s\n", len(booksFromShelf), shelfURL)
	close(stubBooksFoundFromGoodReadsChan)

	searchResults := []dtos.EnchancedSearchResult{}
	for _, book := range booksFromShelf {
		searchResults = append(searchResults, thebookshop.SearchForBook(book, stubSearchResultsFromTheBookshopChan))
	}
	booksFromShelfThatAreAvailableNow = goodreads.GetAvailableBooksFromSearchResult(searchResults)
	logger.Sugar().Infof("%s search queries were made with %d matches found",
		len(searchResults), len(booksFromShelfThatAreAvailableNow))

	newBooksThatNeedNotification := []dtos.AvailableBook{}
	for _, availableBook := range booksFromShelfThatAreAvailableNow {
		if bookIsNew := availableBookIsNew(availableBook, cachedBooksThatAreStillAvailableToday); bookIsNew {
			newBooksThatNeedNotification = append(newBooksThatNeedNotification, availableBook)
		}
	}

	logger.Sugar().Infof("%d new books were found in this search", len(newBooksThatNeedNotification))
	if len(newBooksThatNeedNotification) > 0 {
		for _, newBook := range newBooksThatNeedNotification {
			if authorIsIgnored := db.IsIgnoredAuthor(newBook.BookPurchaseInfo.Author); !authorIsIgnored {
				db.AddAvailableBook(newBook)
				util.SendNewBookIsAvailableMessage(newBook.BookPurchaseInfo)
			}
		}
	}
	logger.Sugar().Infof("%d cached books were available yesterday", len(cachedBooksThatWereAvailable))
	logger.Sugar().Infof("%d books are available today from cache", len(cachedBooksThatAreStillAvailableToday))
	logger.Sugar().Infof("These books are brand new from this current crawl: %+v\n", newBooksThatNeedNotification)

	sendFreeShippingWebhookIfFreeShippingEligible()
}

func Worker(shelfURL string, ws *websocket.Conn) {
	if isValidShelfURL := goodreads.CheckIsShelfURL(shelfURL); !isValidShelfURL {
		errorMsg := fmt.Sprintf("Invalid shelfURL '%s' given", shelfURL)
		writeErrorMsg(errorMsg, ws)
		return
	}

	db.AddNewCrawlBreadcrumb(shelfURL)
	previouslyKnownAvailableBooks := db.GetAvailableBooksMap()

	shelfStatsChan := make(chan int, 1)
	booksFoundFromGoodReadsChan := make(chan dtos.BasicGoodReadsBook, 200)
	searchResultsFromTheBookshopChan := make(chan dtos.EnchancedSearchResult, 200)

	logger.Sugar().Infof("Retrieving books from shelf: %s\n", shelfURL)
	goodreads.GetBooksFromShelf(shelfURL, shelfStatsChan, booksFoundFromGoodReadsChan)

	booksFound := 0
	searchResultsReturned := 0
	totalBooksFromGoodReads := -1
	currCrawlStats := dtos.CrawlStats{
		TotalBooks:    totalBooksFromGoodReads,
		BooksCrawled:  booksFound,
		BooksSearched: searchResultsReturned,
	}
	newBooksFound := 0
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
				writeTotalBooksMsg(currCrawlStats, ws)
			}

		case bookFromGoodReads := <-booksFoundFromGoodReadsChan:
			currCrawlStats.BooksCrawled++
			logger.Sugar().Infof("[booksFound: %d][booksCrawled: %d] Found a GoodReads book: %+v by %v",
				len(booksFoundFromGoodReadsChan), currCrawlStats.BooksCrawled,
				bookFromGoodReads.Title, bookFromGoodReads.Author)
			writeGoodReadsBookMsg(bookFromGoodReads, currCrawlStats, ws)
			go thebookshop.SearchForBook(bookFromGoodReads, searchResultsFromTheBookshopChan)

		case searchResultFromTheBookshop := <-searchResultsFromTheBookshopChan:
			currCrawlStats.BooksSearched++
			currCrawlStats.BookMatchFound += len(searchResultFromTheBookshop.TitleMatches)

			if len(searchResultFromTheBookshop.TitleMatches) > 0 {
				for _, potentialNewBook := range searchResultFromTheBookshop.TitleMatches {
					if authorIsIgnored := db.IsIgnoredAuthor(potentialNewBook.Author); !authorIsIgnored {
						db.AddAuthorToKnownAuthors(potentialNewBook.Author)
						if bookIsNew := goodReadsBookIsNew(potentialNewBook, previouslyKnownAvailableBooks); bookIsNew {
							newBooksFound++
							logger.Sugar().Infof("Found a book that's for sale: %s by %s for %s at %s",
								searchResultFromTheBookshop.SearchBook.Title,
								searchResultFromTheBookshop.SearchBook.Author,
								potentialNewBook.Price,
								potentialNewBook.Link)

							writeNewAvailableBookMsg(potentialNewBook, currCrawlStats, ws)
							newBook := dtos.AvailableBook{
								BookInfo:         searchResultFromTheBookshop.SearchBook,
								BookPurchaseInfo: potentialNewBook,
							}
							db.AddAvailableBook(newBook)
							util.SendNewBookIsAvailableMessage(potentialNewBook)
							previouslyKnownAvailableBooks[potentialNewBook.Link] = true
						}
					}

				}
			}
			if addMoreAuthorBooksToAvailableBooksList := db.GetAddMoreAuthorBooksToAvailableBooksList(); addMoreAuthorBooksToAvailableBooksList {
				if len(searchResultFromTheBookshop.AuthorMatches) > 0 {
					for _, potentialNewBook := range searchResultFromTheBookshop.AuthorMatches {
						if authorIsIgnored := db.IsIgnoredAuthor(potentialNewBook.Author); !authorIsIgnored {
							db.AddAuthorToKnownAuthors(potentialNewBook.Author)
							if bookIsNew := goodReadsBookIsNew(potentialNewBook, previouslyKnownAvailableBooks); bookIsNew {
								newBooksFound++
								currCrawlStats.BookMatchFound++
								logger.Sugar().Infof("Found a book that's for sale: %s by %s for %s at %s",
									potentialNewBook.Title,
									potentialNewBook.Author,
									potentialNewBook.Price,
									potentialNewBook.Link)
								writeNewAvailableBookMsg(potentialNewBook, currCrawlStats, ws)
								newBook := dtos.AvailableBook{
									BookInfo:         searchResultFromTheBookshop.SearchBook,
									BookPurchaseInfo: potentialNewBook,
								}
								db.AddAvailableBook(newBook)
								util.SendNewBookIsAvailableMessage(potentialNewBook)
								previouslyKnownAvailableBooks[potentialNewBook.Link] = true
							}
						}
					}
				}
			}

			writeSearchResultReturnedMsg(searchResultFromTheBookshop, currCrawlStats, ws)
		}
	}

	db.SetTotalBooksInAutomatedBookShelfCheck(currCrawlStats.TotalBooks)

	logger.Sugar().Infof("Finished. Crawled %d books from GoodReads and made %d searches to TheBookshop.ie which had %d new books",
		currCrawlStats.BooksCrawled, currCrawlStats.BooksSearched, newBooksFound)
	close(booksFoundFromGoodReadsChan)
	close(searchResultsFromTheBookshopChan)
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
