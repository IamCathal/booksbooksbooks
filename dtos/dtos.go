package dtos

import "time"

type AppConfig struct {
	ApplicationStartUpTime time.Time
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

// Endpoints DTOs

type UptimeResponse struct {
	Status      string        `json:"status,omitempty"`
	Uptime      time.Duration `json:"uptime,omitempty"`
	StartUpTime int64         `json:"startuptime,omitempty"`
}

type AutomatedShelfCheckURLResponse struct {
	ShelURL string `json:"shelfURL"`
}

type GetDiscordWebhookResponse struct {
	WebHook string `json:"webhook"`
}

type GetDiscordMessageFormatResponse struct {
	Format string `json:"format"`
}

type GetAutomatedCrawlTime struct {
	Time string `json:"time"`
}

// Websocket DTOs

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

type WsNewBookAvailable struct {
	Book       TheBookshopBook `json:"newAvailableBook"`
	CrawlStats CrawlStats      `json:"crawlStats"`
}

// Database related data structures

type RecentCrawl struct {
	CrawlKey string `json:"crawlKey"`
	ShelfURL string `json:"shelfURL"`
}

type AvailableBook struct {
	BookInfo         BasicGoodReadsBook `json:"bookInfo"`
	BookPurchaseInfo TheBookshopBook    `json:"bookPurchaseInfo"`
}

// Discord embed webhook datastructures

type DiscordMsg struct {
	Content    string         `json:"content,omitempty"`
	Username   string         `json:"username,omitempty"`
	Avatar_url string         `json:"avatar_url,omitempty"`
	Embed      []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string       `json:"title,omitempty"`
	EmbedType   string       `json:"type,omitempty"`
	Description string       `json:"description,omitempty"`
	URL         string       `json:"url,omitempty"`
	Timestamp   string       `json:"timestamp,omitempty"`
	Color       int          `json:"color,omitempty"`
	Image       EmbedImage   `json:"image,omitempty"`
	Thumbnail   EmbedImage   `json:"thumbnail,omitempty"`
	Fields      []EmbedField `json:"fields,omitempty"`
	Author      EmbedAuthor  `json:"author,omitempty"`
	Footer      EmbedFooter  `json:"footer,omitempty"`
}

type EmbedAuthor struct {
	Name    string `json:"name,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
	URL     string `json:"url,omitempty"`
}

type EmbedFooter struct {
	Text    string `json:"text,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

type EmbedImage struct {
	URL      string `json:"url,omitempty"`
	ProxyURL string `json:"proxy_url,omitempty"`
	Height   string `json:"height,omitempty"`
	Width    string `json:"width,omitempty"`
}

type EmbedField struct {
	Name   string `json:"name,omitempty"`
	Value  string `json:"value,omitempty"`
	Inline bool   `json:"inline,omitempty"`
}
