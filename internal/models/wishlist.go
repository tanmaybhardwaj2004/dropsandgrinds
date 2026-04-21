package models

type WishlistCreateRequest struct {
	GameID         int64 `json:"game_id"`
	TargetPriceINR int   `json:"target_price_inr"`
}

type WishlistUpdateRequest struct {
	TargetPriceINR int `json:"target_price_inr"`
}

type WishlistItem struct {
	ID              int64  `json:"id" example:"1"`
	UserID          int64  `json:"user_id" example:"1"`
	GameID          int64  `json:"game_id" example:"1"`
	TargetPriceINR  int    `json:"target_price_inr" example:"999"`
	CurrentPriceINR int    `json:"current_price_inr" example:"1499"`
	Triggered       bool   `json:"triggered" example:"false"`
	Title           string `json:"title" example:"Cyberpunk 2077"`
	Platform        string `json:"platform" example:"Steam"`
	CoverURL        string `json:"cover_url" example:"https://example.com/cover.jpg"`
	CreatedAt       string `json:"created_at" example:"2026-04-21T09:30:00Z"`
}

type WishlistListResponse struct {
	Items  []WishlistItem `json:"items"`
	Total  int            `json:"total" example:"2"`
	Limit  int            `json:"limit" example:"20"`
	Offset int            `json:"offset" example:"0"`
}

type BuyAdviceResponse struct {
	GameID            int64  `json:"game_id" example:"1"`
	CurrentPriceINR   int    `json:"current_price_inr" example:"1499"`
	LowestPriceINR    int    `json:"lowest_price_inr" example:"999"`
	AveragePriceINR   int    `json:"average_price_inr" example:"1699"`
	Recommendation    string `json:"recommendation" example:"wait"`
	ConfidencePercent int    `json:"confidence_percent" example:"78"`
	Reason            string `json:"reason" example:"Current price is well above historical low."`
}
