package models

import "time"

type AnalyticsEvent struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id,omitempty"`
	EventType string    `json:"event_type" binding:"required,max=128"`
	EventData string    `json:"event_data" binding:"max=4096"`
	PageURL   string    `json:"page_url" binding:"max=2048"`
	UserAgent string    `json:"user_agent" binding:"max=512"`
	CreatedAt time.Time `json:"created_at"`
}

type AnalyticsEventsRequest struct {
	Events []AnalyticsEvent `json:"events"`
}
