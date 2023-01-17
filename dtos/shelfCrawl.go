package dtos

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

type WsNewBookAvailable struct {
	Book       TheBookshopBook `json:"newAvailableBook"`
	CrawlStats CrawlStats      `json:"crawlStats"`
}

type RecentCrawlBreadcrumb struct {
	CrawlKey     string `json:"crawlKey"`
	ShelfURL     string `json:"shelfURL"`
	BookCount    int    `json:"bookCount"`
	MatchesCount int    `json:"matchesCount"`
}
