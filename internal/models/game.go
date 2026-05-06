package models

type PriceComparison struct {
	Store             string   `json:"store"`
	Region            string   `json:"region"`
	PriceINR          int      `json:"price_inr"`
	OriginalINR       int      `json:"original_inr"`
	DiscountPercent   int      `json:"discount_percent"`
	StoreURL          string   `json:"store_url"`
	PaymentMethods    []string `json:"payment_methods"`
	GSTInclusive      bool     `json:"gst_inclusive"`
	RegionalArbitrage string   `json:"regional_arbitrage"`
}

type Game struct {
	ID              int64  `json:"id" example:"1"`
	Title           string `json:"title" example:"Cyberpunk 2077"`
	Platform        string `json:"platform" example:"Steam"`
	CoverURL        string `json:"cover_url" example:"https://example.com/cover.jpg"`
	BannerURL       string `json:"banner_url,omitempty"`
	StoreURL        string `json:"store_url" example:"https://store.steampowered.com/app/1091500"`
	PriceINR        int    `json:"price_inr" example:"1499"`
	LowestPriceINR  int    `json:"lowest_price_inr" example:"999"`
	IsAllTimeLow    bool   `json:"is_all_time_low" example:"false"`
	OriginalINR     int    `json:"original_inr" example:"2999"`
	DiscountPercent int    `json:"discount_percent" example:"50"`
	ReviewScore     int    `json:"review_score" example:"86"`

	Description        string            `json:"description,omitempty"`
	ReleaseDate        string            `json:"release_date,omitempty"`
	SupportedPlatforms []string          `json:"supported_platforms,omitempty"`
	Genres             []string          `json:"genres,omitempty"`
	Screenshots        []string          `json:"screenshots,omitempty"`
	Trailers           []string          `json:"trailers,omitempty"`
	SystemRequirements map[string]string `json:"system_requirements,omitempty"`
	DLCs               []string          `json:"dlcs,omitempty"`
	Editions           []string          `json:"editions,omitempty"`
	ReviewSummary      string            `json:"review_summary,omitempty"`
	DealQuality        string            `json:"deal_quality,omitempty"`
	GSTAmount          int               `json:"gst_amount,omitempty"`
	TotalWithGST       int               `json:"total_with_gst,omitempty"`
	CheapestRegion     string            `json:"cheapest_region,omitempty"`
	PaymentMethods     []string          `json:"payment_methods,omitempty"`
	PriceComparisons   []PriceComparison `json:"price_comparisons,omitempty"`
	BestTimeToBuy      string            `json:"best_time_to_buy,omitempty"`
	LiveDataSource     string            `json:"live_data_source,omitempty"`
	LastSyncedAt       string            `json:"last_synced_at,omitempty"`
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
	DealQuality         string `json:"deal_quality" example:"hot"`
	PotentialSavingsINR int    `json:"potential_savings_inr" example:"500"`
}

type DealListResponse struct {
	Deals  []Deal `json:"deals"`
	Total  int    `json:"total" example:"8"`
	Limit  int    `json:"limit" example:"20"`
	Offset int    `json:"offset" example:"0"`
}

type PriceHistoryPoint struct {
	PriceINR        int    `json:"price_inr" example:"1499"`
	IsHistoricalLow bool   `json:"is_historical_low" example:"true"`
	FetchedAt       string `json:"fetched_at" example:"2026-04-21T09:30:00Z"`
}

type PriceHistoryResponse struct {
	GameID  int64               `json:"game_id" example:"1"`
	History []PriceHistoryPoint `json:"history"`
	Prices  []PriceHistoryPoint `json:"prices"`
}

type IndiaArbitrage struct {
	GameID           int64  `json:"game_id" example:"1"`
	SteamIndiaPrice  int    `json:"steam_india_price" example:"1499"`
	SteamGlobalPrice int    `json:"steam_global_price" example:"1499"`
	SteamGlobalINR   int    `json:"steam_global_inr" example:"12500"`
	GSTAmount        int    `json:"gst_amount" example:"225"`
	TotalWithGST     int    `json:"total_with_gst" example:"12725"`
	CheapestRegion   string `json:"cheapest_region" example:"India"`
	Verdict          string `json:"verdict" example:"Buy from India - saves ₹11225"`
}
