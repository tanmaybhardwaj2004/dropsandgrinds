package models

type Game struct {
	ID              int64          `json:"id" example:"1"`
	Title           string         `json:"title" example:"Cyberpunk 2077"`
	Platform        string         `json:"platform" example:"Steam"`
	CoverURL        string         `json:"cover_url" example:"https://example.com/cover.jpg"`
	PriceINR        int            `json:"price_inr" example:"1499"`
	LowestPriceINR  int            `json:"lowest_price_inr" example:"999"`
	IsAllTimeLow    bool           `json:"is_all_time_low" example:"false"`
	OriginalINR     int            `json:"original_inr" example:"2999"`
	DiscountPercent int            `json:"discount_percent" example:"50"`
	ReviewScore     int            `json:"review_score" example:"86"`
	Arbitrage       *ArbitrageData `json:"arbitrage,omitempty"`
}

type GameListResponse struct {
	Games  []Game `json:"games"`
	Total  int    `json:"total" example:"8"`
	Limit  int    `json:"limit" example:"20"`
	Offset int    `json:"offset" example:"0"`
}

type Deal struct {
	Game
	CachedAt            string `json:"cached_at" example:"2026-04-21T09:30:00Z"`
	DealStatus          string `json:"deal_status" example:"good"`
	PotentialSavingsINR int    `json:"potential_savings_inr" example:"500"`
}

type DealListResponse struct {
	Deals  []Deal `json:"deals"`
	Total  int    `json:"total" example:"8"`
	Limit  int    `json:"limit" example:"20"`
	Offset int    `json:"offset" example:"0"`
}

type PriceHistoryPoint struct {
	PriceINR  int    `json:"price_inr" example:"1499"`
	FetchedAt string `json:"fetched_at" example:"2026-04-21T09:30:00Z"`
}

type PriceHistoryResponse struct {
	GameID  int64               `json:"game_id" example:"1"`
	History []PriceHistoryPoint `json:"history"`
}

type ArbitrageData struct {
	IndiaBaseINR   float64 `json:"india_base_inr" example:"1499"`
	IndiaGSTINR    float64 `json:"india_gst_inr" example:"270"`
	IndiaTotalINR  float64 `json:"india_total_inr" example:"1769"`
	GlobalBaseINR  float64 `json:"global_base_inr" example:"12500"`
	GlobalGSTINR   float64 `json:"global_gst_inr" example:"2250"`
	GlobalTotalINR float64 `json:"global_total_inr" example:"14750"`
	CheapestRegion string  `json:"cheapest_region" example:"India"`
	Verdict        string  `json:"verdict" example:"Buy from India - saves ₹12981"`
}
