package dtos

type SeriesCrawlStats struct {
	BooksInShelf               int `json:"booksInShelf"`
	SeriesCount                int `json:"seriesCount"`
	TotalBooksInSeries         int `json:"totalBooksInSeries"`
	BooksSearchedOnTheBookshop int `json:"booksSearchedOnTheBookshop"`
	SeriesLookedUp             int `json:"seriesLookedUp"`
	BookMatchesFound           int `json:"bookMatchesFound"`
}

type Series struct {
	ID           string       `json:"id"`
	Author       string       `json:"author"`
	Title        string       `json:"title"`
	PrimaryWorks int          `json:"primaryWorks"`
	TotalWorks   int          `json:"totalWorks"`
	Link         string       `json:"link"`
	Books        []SeriesBook `json:"books"`
	Ignore       bool         `json:"ignore"`
}

type SeriesBook struct {
	BookSeriesText   string             `json:"bookSeriesText"`
	RealBookOrder    int                `json:"realBookOrder"`
	BookInfo         BasicGoodReadsBook `json:"bookInfo"`
	TheBookshopMatch TheBookshopBook    `json:"theBookshopMatch"`
}

type NewBookInSeries struct {
	SeriesBaseTitle string `json:"seriesBaseTitle"`
}

type WsSeriesCrawlStats struct {
	CrawlStats SeriesCrawlStats `json:"seriesCrawlStats"`
}

type WsNewSeries struct {
	Series     Series           `json:"series"`
	CrawlStats SeriesCrawlStats `json:"seriesCrawlStats"`
}

type WsSearchResultReturned struct {
	SearchBook BasicGoodReadsBook `json:"searchBook"`
	Match      TheBookshopBook    `json:"match"`
	CrawlStats SeriesCrawlStats   `json:"seriesCrawlStats"`
}
