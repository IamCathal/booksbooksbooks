package db

import (
	"encoding/json"
	"time"

	redis "github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	AVAILABLE_BOOKS          = "availableBooks"
	SERIES_CRAWL_BOOKS       = "seriesCrawlBooks"
	RECENT_CRAWL_BREADCRUMBS = "recentCrawlBreadcrumbs"
	RECENT_CRAWL_REPORTS     = "recentCrawlReports"
)

func ResetAvailableBooks() {
	err := redisClient.Set(ctx, AVAILABLE_BOOKS, []byte(""), DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func AddAvailableBook(newBook dtos.AvailableBook) {
	availableBooks := GetAvailableBooks()
	newBook.Ignore = false
	newBook.LastCheckedTimeStamp = time.Now().Unix()
	availableBooks = append(availableBooks, newBook)

	availableBooksWithNoDuplicates := removeDuplicateAvailableBooks(availableBooks)
	jsonAvailableBooks, err := json.Marshal(availableBooksWithNoDuplicates)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	err = redisClient.Set(ctx, AVAILABLE_BOOKS, jsonAvailableBooks, time.Duration(DEFAULT_TTL)).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func SetAvailableBooks(availableBooks []dtos.AvailableBook) {
	availableBooksJson, err := json.Marshal(availableBooks)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, AVAILABLE_BOOKS, availableBooksJson, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAvailableBooks() []dtos.AvailableBook {
	availableBooksStr, err := redisClient.Get(ctx, AVAILABLE_BOOKS).Result()
	if err == redis.Nil {
		return []dtos.AvailableBook{}
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	availableBooks := []dtos.AvailableBook{}
	if availableBooksStr != "" {
		err = json.Unmarshal([]byte(availableBooksStr), &availableBooks)
		if err != nil {
			logger.Sugar().Fatal(err)
		}
	}
	return availableBooks
}

func RemoveAvailableBook(bookToRemove dtos.AvailableBook) {
	updatedAvailableBooks := []dtos.AvailableBook{}

	for _, currBook := range GetAvailableBooks() {
		if bookToRemove.BookInfo.ID != currBook.BookInfo.ID {
			updatedAvailableBooks = append(updatedAvailableBooks, currBook)
		}
	}

	SetAvailableBooks(updatedAvailableBooks)
}

func SetSeriesCrawlBooks(seriesCrawlBooks []dtos.Series) {
	seriesCrawlBooksJson, err := json.Marshal(seriesCrawlBooks)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, SERIES_CRAWL_BOOKS, seriesCrawlBooksJson, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetSeriesCrawlBooks() []dtos.Series {
	seriesCrawlBooksStr, err := redisClient.Get(ctx, SERIES_CRAWL_BOOKS).Result()
	if err == redis.Nil {
		return []dtos.Series{}
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	seriesCrawlBooks := []dtos.Series{}
	if seriesCrawlBooksStr != "" {
		err = json.Unmarshal([]byte(seriesCrawlBooksStr), &seriesCrawlBooks)
		if err != nil {
			logger.Sugar().Fatal(err)
		}
	}
	return seriesCrawlBooks
}

func GetRecentCrawlBreadcrumbs() []dtos.RecentCrawlBreadcrumb {
	recentCrawlBreadcrumbs, err := redisClient.Get(ctx, RECENT_CRAWL_BREADCRUMBS).Result()
	if err == redis.Nil {
		return []dtos.RecentCrawlBreadcrumb{}
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	recentCrawlBreadcrumbsArr := []dtos.RecentCrawlBreadcrumb{}
	if recentCrawlBreadcrumbs != "" {
		err = json.Unmarshal([]byte(recentCrawlBreadcrumbs), &recentCrawlBreadcrumbsArr)
		if err != nil {
			logger.Sugar().Fatal(err)
		}
	}
	return recentCrawlBreadcrumbsArr
}

func AddNewCrawlBreadcrumb(shelfURL string, bookCount int) {
	recentCrawlBreadcrumbs := GetRecentCrawlBreadcrumbs()

	updatedCrawlBreadcrumbs := []dtos.RecentCrawlBreadcrumb{
		{
			CrawlKey:  GetKeyForRecentCrawlBreadcrumb(shelfURL),
			ShelfURL:  shelfURL,
			BookCount: bookCount,
		},
	}
	logger.Sugar().Infof("Creating new crawl breadcrumb with key: %s for shelfURL: %s",
		updatedCrawlBreadcrumbs[0].CrawlKey, updatedCrawlBreadcrumbs[0].ShelfURL)

	updatedCrawlBreadcrumbs = append(updatedCrawlBreadcrumbs, recentCrawlBreadcrumbs...)
	updatedCrawlBreadcrumbs = removeDuplicateRecentCrawls(updatedCrawlBreadcrumbs)

	jsonCrawlBreadcrumbs, err := json.Marshal(updatedCrawlBreadcrumbs)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, RECENT_CRAWL_BREADCRUMBS, jsonCrawlBreadcrumbs, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func SetRecentCrawlBreadcrumbs(breadCrumbs []dtos.RecentCrawlBreadcrumb) {
	jsonCrawlBreadcrumbs, err := json.Marshal(breadCrumbs)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, RECENT_CRAWL_BREADCRUMBS, jsonCrawlBreadcrumbs, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetRecentCrawlReports() []dtos.AutomatedCrawlReport {
	recentCrawlReportsStr, err := redisClient.Get(ctx, RECENT_CRAWL_REPORTS).Result()
	if err == redis.Nil {
		SetRecentCrawlReports([]dtos.AutomatedCrawlReport{})
		return GetRecentCrawlReports()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	recentCrawlReportsArr := []dtos.AutomatedCrawlReport{}
	if recentCrawlReportsStr == "" {
		SetRecentCrawlReports([]dtos.AutomatedCrawlReport{})
		return GetRecentCrawlReports()
	}

	err = json.Unmarshal([]byte(recentCrawlReportsStr), &recentCrawlReportsArr)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	return recentCrawlReportsArr
}

func SetRecentCrawlReports(crawlReports []dtos.AutomatedCrawlReport) {
	jsonCrawlReports, err := json.Marshal(crawlReports)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, RECENT_CRAWL_REPORTS, jsonCrawlReports, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func AddNewRecentCrawlReport(crawlReport dtos.AutomatedCrawlReport) {
	recentCrawlReports := GetRecentCrawlReports()
	newCrawlReport := []dtos.AutomatedCrawlReport{
		crawlReport,
	}
	allCrawlReports := append(newCrawlReport, recentCrawlReports...)
	if len(allCrawlReports) >= 15 {
		allCrawlReports = allCrawlReports[:15]
	}
	SetRecentCrawlReports(allCrawlReports)
}
