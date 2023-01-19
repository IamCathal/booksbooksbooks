package dtos

import "time"

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

type GetAutomatedCrawlTimeResponse struct {
	Time string `json:"time"`
}

type BooleanSettingStatusResponse struct {
	Enabled bool `json:"enabled"`
}

type GetPreviewForShelfResponse struct {
	ShelfToCrawl ShelfToCrawl `json:"shelfToCrawlPreview"`
}

type GetPreviewsForShelvesResponse struct {
	ShelvesToCrawl []ShelfToCrawl `json:"shelvesToCrawlPreviews"`
}

type GetKnownAuthorsResponse struct {
	Authors []string `json:"authors"`
}
