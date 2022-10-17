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
	Title      string `json:"title"`
	Author     string `json:"author"`
	SeriesText string `json:"seriesText"`
}

type TheBookshopBook struct {
}
