package dtos

import "time"

type AppConfig struct {
	ApplicationStartUpTime time.Time
}

type UptimeResponse struct {
	Status      string        `json:"status,omitempty"`
	Uptime      time.Duration `json:"uptime,omitempty"`
	StartUpTime int64         `json:"startuptime,omitempty"`
}

type BasicGoodReadsBook struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Author     string  `json:"author"`
	SeriesText string  `json:"seriesText"`
	Link       string  `json:"link"`
	Cover      string  `json:"cover"`
	Isbn13     string  `json:"isbn13"`
	Asin       string  `json:"asin"`
	Rating     float32 `json:"rating"`
}

type TheBookshopBook struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Price  string `json:"price"`
	Link   string `json:"link"`
	Cover  string `json:"cover"`
}

type BookShopBookSearchResult struct {
	SearchResultBooks []TheBookshopBook `json:"searchResultBooks"`
}

type AllBookshopBooksSearchResults map[string]BookShopBookSearchResult

type EnchancedSearchResult struct {
	SearchBook    BasicGoodReadsBook `json:"searchBook"`
	AuthorMatches []TheBookshopBook  `json:"authorMatches"`
	TitleMatches  []TheBookshopBook  `json:"titleMatches"`
}

// Ws data structures

type WsErrorMsg struct {
	Error string `json:"error"`
}

type CrawlStats struct {
	TotalBooks     int `json:"totalBooks"`
	BooksCrawled   int `json:"booksCrawled"`
	BooksSearched  int `json:"booksSearched"`
	BookMatchFound int `json:"bookMatchFound"`
}

type WsTotalBooks struct {
	TotalBooks int        `json:"totalBooks"`
	CrawlStats CrawlStats `json:"crawlStats"`
}

type WsGoodreadsBook struct {
	BookInfo   BasicGoodReadsBook `json:"bookinfo"`
	CrawlStats CrawlStats         `json:"crawlStats"`
}

type WsBookshopSearchResult struct {
	SearchResult EnchancedSearchResult `json:"searchResult"`
	CrawlStats   CrawlStats            `json:"crawlStats"`
}
