package db

import (
	"encoding/json"

	redis "github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	AUTOMATED_BOOK_SHELF_CHECK_URL                = "automatedBookShelfCheck"
	AUTOMATED_BOOK_SHELF_CRAWL_TIME               = "automatedBookShelfCrawlTime"
	DISCORD_WEBHOOK_URL                           = "discordWebHookURL"
	DISCORD_MESSAGE_FORMAT                        = "discordMessageFormat"
	SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE      = "sendAlertWhenBookNoLongerAvailable"
	SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN   = "sendAlertWhenFreeShippingKicksIn"
	TOTAL_BOOKS_IN_AUTOMATED_BOOK_SHELF           = "totalBooksInAutomatedBookShelf"
	ADD_MORE_AUTHOR_BOOKS_TO_AVAILABLE_BOOKS_LIST = "addMoreAuthorBooksToAvailableBooksList"
	KNOWN_AUTHORS                                 = "knownAuthors"
	IGNORE_AUTHORS                                = "ignoreAuthors"
	OTHER_BOOKS_IN_SERIES_LOOKUP                  = "otherBooksInSeriesLookup"
	OWNED_BOOKS_SHELF_URL                         = "ownedBooksShelfURL"
	ONLY_ENGLISH_BOOKS_TOGGLE                     = "onlyEnglishBooksToggle"
	SERIES_CRAWL_IN_AUTOMATED_CRAWL               = "seriesCrawlInAutomatedCrawl"
)

func SetAutomatedBookShelfCheck(shelfURL string) {
	err := redisClient.Set(ctx, AUTOMATED_BOOK_SHELF_CHECK_URL, shelfURL, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAutomatedBookShelfCheck() string {
	shelfURL, err := redisClient.Get(ctx, AUTOMATED_BOOK_SHELF_CHECK_URL).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return shelfURL
}

func SetDiscordWebhookURL(webhookURL string) {
	err := redisClient.Set(ctx, DISCORD_WEBHOOK_URL, webhookURL, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetDiscordWebhookURL() string {
	webhookURL, err := redisClient.Get(ctx, DISCORD_WEBHOOK_URL).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return webhookURL
}

func SetDiscordMessageFormat(format string) {
	err := redisClient.Set(ctx, DISCORD_MESSAGE_FORMAT, format, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetDiscordMessageFormat() string {
	format, err := redisClient.Get(ctx, DISCORD_MESSAGE_FORMAT).Result()
	if err == redis.Nil {
		SetDiscordMessageFormat("small")
		return GetDiscordMessageFormat()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if format == "" {
		SetDiscordMessageFormat("small")
		return GetDiscordMessageFormat()
	}
	return format
}

func SetAutomatedBookShelfCrawlTime(time string) {
	err := redisClient.Set(ctx, AUTOMATED_BOOK_SHELF_CRAWL_TIME, time, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAutomatedBookShelfCrawlTime() string {
	time, err := redisClient.Get(ctx, AUTOMATED_BOOK_SHELF_CRAWL_TIME).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return time
}

func SetSendAlertWhenBookNoLongerAvailable(enabled bool) {
	err := redisClient.Set(ctx, SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetSendAlertWhenBookNoLongerAvailable() bool {
	enabled, err := redisClient.Get(ctx, SEND_ALERT_WHEN_BOOK_NO_LONGER_AVAILABLE).Result()
	if err == redis.Nil {
		SetSendAlertWhenBookNoLongerAvailable(false)
		return GetSendAlertWhenBookNoLongerAvailable()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetSendAlertWhenBookNoLongerAvailable(false)
		return GetSendAlertWhenBookNoLongerAvailable()
	}
	return strToBool(enabled)
}

func SetSendAlertOnlyWhenFreeShippingKicksIn(enabled bool) {
	err := redisClient.Set(ctx, SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetSendAlertOnlyWhenFreeShippingKicksIn() bool {
	enabled, err := redisClient.Get(ctx, SEND_ALERT_ONLY_WHEN_FREE_SHIPPING_KICKS_IN).Result()
	if err == redis.Nil {
		SetSendAlertOnlyWhenFreeShippingKicksIn(false)
		return GetSendAlertOnlyWhenFreeShippingKicksIn()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetSendAlertOnlyWhenFreeShippingKicksIn(false)
		return GetSendAlertOnlyWhenFreeShippingKicksIn()
	}
	return strToBool(enabled)
}

func SetTotalBooksInAutomatedBookShelfCheck(totalBooks int) {
	err := redisClient.Set(ctx, TOTAL_BOOKS_IN_AUTOMATED_BOOK_SHELF, totalBooks, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetTotalBooksInAutomatedBookShelfCheck() int {
	totalBooks, err := redisClient.Get(ctx, TOTAL_BOOKS_IN_AUTOMATED_BOOK_SHELF).Result()
	if err == redis.Nil {
		return 0
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return strToInt(totalBooks)
}

func SetAddMoreAuthorBooksToAvailableBooksList(enabled bool) {
	err := redisClient.Set(ctx, ADD_MORE_AUTHOR_BOOKS_TO_AVAILABLE_BOOKS_LIST, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetAddMoreAuthorBooksToAvailableBooksList() bool {
	enabled, err := redisClient.Get(ctx, ADD_MORE_AUTHOR_BOOKS_TO_AVAILABLE_BOOKS_LIST).Result()
	if err == redis.Nil {
		SetAddMoreAuthorBooksToAvailableBooksList(false)
		return GetAddMoreAuthorBooksToAvailableBooksList()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetAddMoreAuthorBooksToAvailableBooksList(false)
		return GetAddMoreAuthorBooksToAvailableBooksList()
	}
	return strToBool(enabled)
}

func GetKnownAuthors() []dtos.KnownAuthor {
	knownAuthors, err := redisClient.Get(ctx, KNOWN_AUTHORS).Result()
	if err == redis.Nil {
		return []dtos.KnownAuthor{}
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	knownAuthorsArr := []dtos.KnownAuthor{}
	if knownAuthors != "" {
		err = json.Unmarshal([]byte(knownAuthors), &knownAuthorsArr)
		if err != nil {
			logger.Sugar().Fatal(err)
		}
	}
	return knownAuthorsArr
}

func getIgnoredAuthors() []string {
	knownAuthors := GetKnownAuthors()
	ignoredAuthors := []string{}
	for _, author := range knownAuthors {
		if author.Ignore {
			ignoredAuthors = append(ignoredAuthors, author.Name)
		}
	}
	return ignoredAuthors
}

func AddAuthorToKnownAuthors(author string) {
	knownAuthors := GetKnownAuthors()

	knownAuthors = append(knownAuthors, dtos.KnownAuthor{Name: author, Ignore: false})
	knownAuthors = removeDuplicateAuthors(knownAuthors)
	knownAuthors = removeDuplicateAuthorsDisregardingReverseOrder(knownAuthors)
	jsonKnownAuthors, err := json.Marshal(knownAuthors)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	err = redisClient.Set(ctx, KNOWN_AUTHORS, jsonKnownAuthors, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func SetKnownAuthors(authors []dtos.KnownAuthor) {
	jsonKnownAuthors, err := json.Marshal(authors)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	err = redisClient.Set(ctx, KNOWN_AUTHORS, jsonKnownAuthors, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func ToggleAuthorIgnore(authorToSearch string) {
	knownAuthors := GetKnownAuthors()
	newKnownAuthors := []dtos.KnownAuthor{}

	for _, author := range knownAuthors {
		if author.Name == authorToSearch {
			if author.Ignore {
				author.Ignore = false
			} else {
				author.Ignore = true
			}
		}
		newKnownAuthors = append(newKnownAuthors, author)
	}
	SetKnownAuthors(newKnownAuthors)
}

func PurgeIgnoredAuthorsFromAvailableBooks() {
	availableBooksWithoutPurgedAuthor := []dtos.AvailableBook{}
	availableBooks := GetAvailableBooks()
	knownAuthors := GetKnownAuthors()

	ignoredAuthorsMap := make(map[string]bool)
	for _, author := range knownAuthors {
		if author.Ignore {
			ignoredAuthorsMap[author.Name] = true
		}
	}

	for _, book := range availableBooks {
		if _, isIgnoredAuthor := ignoredAuthorsMap[book.BookPurchaseInfo.Author]; !isIgnoredAuthor {
			availableBooksWithoutPurgedAuthor = append(availableBooksWithoutPurgedAuthor, book)
		}
	}
	SetAvailableBooks(availableBooksWithoutPurgedAuthor)
}

func SetOtherBooksInSeriesLookup(enabled bool) {
	err := redisClient.Set(ctx, OTHER_BOOKS_IN_SERIES_LOOKUP, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetOtherBooksInSeriesLookup() bool {
	enabled, err := redisClient.Get(ctx, OTHER_BOOKS_IN_SERIES_LOOKUP).Result()
	if err == redis.Nil {
		SetOtherBooksInSeriesLookup(false)
		return GetOtherBooksInSeriesLookup()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetOtherBooksInSeriesLookup(false)
		return GetOtherBooksInSeriesLookup()
	}
	return strToBool(enabled)
}

func SetOwnedBooksShelfURL(shelfURL string) {
	err := redisClient.Set(ctx, OWNED_BOOKS_SHELF_URL, shelfURL, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetOwnedBooksShelfURL() string {
	shelfURL, err := redisClient.Get(ctx, OWNED_BOOKS_SHELF_URL).Result()
	if err == redis.Nil {
		return ""
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	return shelfURL
}

func SetOnlyEnglishBooks(enabled bool) {
	err := redisClient.Set(ctx, ONLY_ENGLISH_BOOKS_TOGGLE, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetOnlyEnglishBooks() bool {
	enabled, err := redisClient.Get(ctx, ONLY_ENGLISH_BOOKS_TOGGLE).Result()
	if err == redis.Nil {
		SetOnlyEnglishBooks(false)
		return GetOnlyEnglishBooks()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetOnlyEnglishBooks(false)
		return GetOnlyEnglishBooks()
	}
	return strToBool(enabled)
}

func SetSeriesCrawlInAutomatedCrawl(enabled bool) {
	err := redisClient.Set(ctx, SERIES_CRAWL_IN_AUTOMATED_CRAWL, enabled, DEFAULT_TTL).Err()
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func GetSeriesCrawlInAutomatedCrawl() bool {
	enabled, err := redisClient.Get(ctx, SERIES_CRAWL_IN_AUTOMATED_CRAWL).Result()
	if err == redis.Nil {
		SetSeriesCrawlInAutomatedCrawl(false)
		return GetSeriesCrawlInAutomatedCrawl()
	} else if err != nil {
		logger.Sugar().Fatal(err)
	}
	if enabled == "" {
		SetSeriesCrawlInAutomatedCrawl(false)
		return GetSeriesCrawlInAutomatedCrawl()
	}
	return strToBool(enabled)
}
