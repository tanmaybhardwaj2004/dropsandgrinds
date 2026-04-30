package models

import "time"

type AnalyticsEvent struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id,omitempty"`
	EventType string    `json:"event_type"`
	EventData string    `json:"event_data"`
	PageURL   string    `json:"page_url"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

type AnalyticsEventsRequest struct {
	Events []AnalyticsEvent `json:"events"`
}
