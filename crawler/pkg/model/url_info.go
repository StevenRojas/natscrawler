package model

import "time"

type UrlInfo struct {
	RequestID   string  `json:"request_id"`
	Url         string  `json:"url"`
	AppName     string  `json:"app_name"`
	Rating      float64 `json:"rating"`
	RatingCount int     `json:"rating_count"`
	Success     bool    `json:"success"`
	LastError   string  `json:"last_error"`
	Stats       Stats   `json:"stats"`
}

type Stats struct {
	CollectorID int   `json:"collector_id"`
	Waiting     Times `json:"waiting"`
	Collector   Times `json:"collector"`
}

type Times struct {
	StartAt  time.Time `json:"start_at"`
	EndAt    time.Time `json:"end_at"`
	Duration int64     `json:"duration"`
}
