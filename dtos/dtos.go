package dtos

import "time"

const (
	TITLE_MATCH = iota
	AUTHOR_MATCH
	SERIES_MATCH
)

type AppConfig struct {
	ApplicationStartUpTime time.Time
}

type EnchancedSearchResult struct {
	SearchBook    BasicGoodReadsBook `json:"searchBook"`
	AuthorMatches []TheBookshopBook  `json:"authorMatches"`
	TitleMatches  []TheBookshopBook  `json:"titleMatches"`
}

type WsErrorMsg struct {
	Error string `json:"error"`
}

type AvailableBook struct {
	BookInfo             BasicGoodReadsBook `json:"bookInfo"`
	BookPurchaseInfo     TheBookshopBook    `json:"bookPurchaseInfo"`
	BookFoundFrom        int                `json:"bookFoundFrom"`
	Ignore               bool               `json:"ignore"`
	LastCheckedTimeStamp int64              `json:"lastCheckedTimeStamp"`
}

type KnownAuthor struct {
	Name   string `json:"name"`
	Ignore bool   `json:"ignore"`
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
