package models

import "time"

// Enhanced Game model with comprehensive multi-platform support
type EnhancedGame struct {
	ID              int64          `json:"id" example:"1"`
	ExternalID      string         `json:"external_id" example:"12345"`
	Title           string         `json:"title" example:"Cyberpunk 2077"`
	Slug            string         `json:"slug" example:"cyberpunk-2077"`
	Description     string         `json:"description" example:"An open-world action-adventure game"`
	ReleaseDate     *time.Time     `json:"release_date" example:"2020-12-10T00:00:00Z"`
	Developer       string         `json:"developer" example:"CD Projekt Red"`
	Publisher       string         `json:"publisher" example:"CD Projekt Red"`
	Genres          []string       `json:"genres" example:"[\"RPG\",\"Action\"]"`
	Platforms       []string       `json:"platforms" example:"[\"PC\",\"PlayStation\",\"Xbox\"]"`
	CoverURL        string         `json:"cover_url" example:"https://example.com/cover.jpg"`
	Screenshots     []string       `json:"screenshots" example:"[\"https://example.com/screen1.jpg\"]"`
	Trailers        []string       `json:"trailers" example:"[\"https://example.com/trailer.mp4\"]"`
	SystemRequirements *SystemReqs    `json:"system_requirements,omitempty"`
	Editions        []Edition      `json:"editions,omitempty"`
	DLCInfo         *DLCInfo        `json:"dlc_info,omitempty"`
	Rating          string         `json:"rating" example:"M"`
	UserRating      float64        `json:"user_rating" example:"8.5"`
	IsDLC           bool           `json:"is_dlc" example:"false"`
	ParentGameID    *int64         `json:"parent_game_id,omitempty"`
	Region          string         `json:"region" example:"IN"`
	IsActive        bool           `json:"is_active" example:"true"`
	LastPriceUpdate *time.Time     `json:"last_price_update,omitempty"`
}

// Store model for multi-platform support
type Store struct {
	ID                int64     `json:"id" example:"1"`
	Name              string     `json:"name" example:"Steam"`
	Slug              string     `json:"slug" example:"steam"`
	WebsiteURL        string     `json:"website_url" example:"https://store.steampowered.com"`
	APIEndpoint       *string    `json:"api_endpoint,omitempty"`
	APIKeyRequired    bool       `json:"api_key_required" example:"true"`
	SupportsIndia    bool       `json:"supports_india" example:"true"`
	Region            string     `json:"region" example:"IN"`
	Currency          string     `json:"currency" example:"INR"`
	ConversionRate    float64    `json:"conversion_rate_to_inr" example:"1.0"`
	LogoURL           *string    `json:"logo_url,omitempty"`
	IsActive          bool       `json:"is_active" example:"true"`
	CreatedAt         time.Time  `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt         time.Time  `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

// Enhanced Price model with store and region support
type EnhancedPrice struct {
	ID             int64     `json:"id" example:"1"`
	GameID         int64     `json:"game_id" example:"1"`
	StoreID        int64     `json:"store_id" example:"1"`
	Store          Store      `json:"store"`
	ExternalID     string     `json:"external_id" example:"12345"`
	PriceINR       float64    `json:"price_inr" example:"1499.99"`
	OriginalPrice  float64    `json:"original_price" example:"2999.99"`
	DiscountAmount float64    `json:"discount_amount" example:"1500.00"`
	DiscountPercent int        `json:"discount_percent" example:"50"`
	Region         string     `json:"region" example:"IN"`
	Currency       string     `json:"currency" example:"INR"`
	IsAvailable    bool       `json:"is_available" example:"true"`
	StockStatus    string     `json:"stock_status" example:"in_stock"`
	DealType       string     `json:"deal_type" example:"regular"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

// Price History model for tracking
type PriceHistory struct {
	ID             int64     `json:"id" example:"1"`
	GameID         int64     `json:"game_id" example:"1"`
	StoreID        int64     `json:"store_id" example:"1"`
	Store          Store      `json:"store"`
	ExternalID     string     `json:"external_id" example:"12345"`
	PriceINR       float64    `json:"price_inr" example:"1499.99"`
	OriginalPrice  float64    `json:"original_price" example:"2999.99"`
	DiscountPercent int        `json:"discount_percent" example:"50"`
	Region         string     `json:"region" example:"IN"`
	Currency       string     `json:"currency" example:"INR"`
	IsAvailable    bool       `json:"is_available" example:"true"`
	RecordedAt     time.Time  `json:"recorded_at" example:"2026-01-01T00:00:00Z"`
}

// Game Metadata model for additional information
type GameMetadata struct {
	ID        int64     `json:"id" example:"1"`
	GameID    int64     `json:"game_id" example:"1"`
	Key       string     `json:"key" example:"min_requirements"`
	Value     string     `json:"value" example:"Intel i5-3470"`
	DataType  string     `json:"data_type" example:"text"`
	CreatedAt time.Time  `json:"created_at" example:"2026-01-01T00:00:00Z"`
}

// Store Integration model for API configurations
type StoreIntegration struct {
	ID               int64      `json:"id" example:"1"`
	StoreID          int64       `json:"store_id" example:"1"`
	IntegrationType   string      `json:"integration_type" example:"api"`
	Config           interface{} `json:"config,omitempty"`
	IsActive         bool        `json:"is_active" example:"true"`
	LastSync         *time.Time  `json:"last_sync,omitempty"`
	SyncFrequency    int         `json:"sync_frequency" example:"3600"`
	CreatedAt        time.Time   `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt        time.Time   `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

// Regional Pricing model
type RegionalPricing struct {
	ID             int64     `json:"id" example:"1"`
	GameID         int64     `json:"game_id" example:"1"`
	StoreID        int64     `json:"store_id" example:"1"`
	Store          Store      `json:"store"`
	Region         string     `json:"region" example:"IN"`
	Currency       string     `json:"currency" example:"INR"`
	Price          float64    `json:"price" example:"1499.99"`
	OriginalPrice  float64    `json:"original_price" example:"2999.99"`
	DiscountPercent int        `json:"discount_percent" example:"50"`
	IsAvailable    bool       `json:"is_available" example:"true"`
	UpdatedAt      time.Time  `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

// Deal Alert model for wishlist notifications
type DealAlert struct {
	ID              int64      `json:"id" example:"1"`
	UserID          int64       `json:"user_id" example:"1"`
	GameID          int64       `json:"game_id" example:"1"`
	StoreID         *int64      `json:"store_id,omitempty"`
	TargetPrice     float64     `json:"target_price" example:"999.99"`
	Region          string       `json:"region" example:"IN"`
	Currency        string       `json:"currency" example:"INR"`
	IsActive        bool         `json:"is_active" example:"true"`
	NotificationSent bool         `json:"notification_sent" example:"false"`
	CreatedAt       time.Time    `json:"created_at" example:"2026-01-01T00:00:00Z"`
	TriggeredAt     *time.Time   `json:"triggered_at,omitempty"`
}

// Trending Deal model
type TrendingDeal struct {
	ID            int64     `json:"id" example:"1"`
	GameID        int64     `json:"game_id" example:"1"`
	StoreID       int64     `json:"store_id" example:"1"`
	Store         Store      `json:"store"`
	TrendScore    float64    `json:"trend_score" example:"85.5"`
	ViewCount     int        `json:"view_count" example:"1500"`
	ClickCount    int        `json:"click_count" example:"250"`
	ConversionRate float64    `json:"conversion_rate" example:"16.7"`
	TrendPeriod   string     `json:"trend_period" example:"24h"`
	IsActive      bool       `json:"is_active" example:"true"`
	CreatedAt     time.Time  `json:"created_at" example:"2026-01-01T00:00:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" example:"2026-01-01T00:00:00Z"`
}

// Indian Payment Offer model
type IndianPaymentOffer struct {
	ID                int64      `json:"id" example:"1"`
	StoreID           int64       `json:"store_id" example:"1"`
	Store             Store       `json:"store"`
	OfferType         string      `json:"offer_type" example:"upi_discount"`
	Provider          string      `json:"provider" example:"PhonePe"`
	Description       string      `json:"description" example:"10% cashback on UPI payments"`
	DiscountPercent    int         `json:"discount_percent" example:"10"`
	MaxDiscountAmount float64     `json:"max_discount_amount" example:"100.00"`
	MinOrderAmount   float64     `json:"min_order_amount" example:"500.00"`
	ValidFrom         *time.Time  `json:"valid_from,omitempty"`
	ValidUntil        *time.Time  `json:"valid_until,omitempty"`
	IsActive          bool        `json:"is_active" example:"true"`
	CreatedAt         time.Time   `json:"created_at" example:"2026-01-01T00:00:00Z"`
}

// System Requirements model
type SystemReqs struct {
	OS          string                 `json:"os" example:"Windows 10 64-bit"`
	Processor   string                 `json:"processor" example:"Intel Core i7-6700"`
	Memory      string                 `json:"memory" example:"16 GB RAM"`
	Graphics    string                 `json:"graphics" example:"NVIDIA GeForce GTX 1060"`
	DirectX    string                 `json:"directx" example:"12"`
	Storage     string                 `json:"storage" example:"70 GB HDD"`
	Additional  map[string]interface{}  `json:"additional,omitempty"`
}

// Edition model for game editions
type Edition struct {
	ID           int64   `json:"id" example:"1"`
	Name         string   `json:"name" example:"Standard Edition"`
	Description   string   `json:"description" example:"Base game"`
	PriceINR     float64  `json:"price_inr" example:"1499.99"`
	IsDLC        bool     `json:"is_dlc" example:"false"`
	Features      []string `json:"features" example:"[\"Base Game\",\"Manual\"]"`
	ReleaseDate   *time.Time `json:"release_date,omitempty"`
}

// DLC Info model
type DLCInfo struct {
	TotalDLCs     int              `json:"total_dlcs" example:"5"`
	AvailableDLCs int              `json:"available_dlcs" example:"3"`
	DLCNames       []string         `json:"dlc_names" example:"[\"Expansion Pack\",\"Season Pass\"]"`
	TotalPriceINR  float64          `json:"total_price_inr" example:"2999.99"`
	BundlesAvailable bool             `json:"bundles_available" example:"true"`
}

// Response models for API endpoints
type EnhancedGameListResponse struct {
	Games  []EnhancedGame `json:"games"`
	Total   int            `json:"total" example:"100"`
	Limit   int            `json:"limit" example:"20"`
	Offset  int            `json:"offset" example:"0"`
}

type PriceComparisonResponse struct {
	Game           EnhancedGame     `json:"game"`
	Prices         []EnhancedPrice  `json:"prices"`
	LowestPrice    *EnhancedPrice   `json:"lowest_price,omitempty"`
	RegionalPrices []RegionalPricing `json:"regional_prices,omitempty"`
	PriceHistory   []PriceHistory  `json:"price_history,omitempty"`
	LastUpdated    time.Time        `json:"last_updated" example:"2026-01-01T00:00:00Z"`
}

type TrendingDealsResponse struct {
	Deals  []TrendingDeal `json:"deals"`
	Total   int            `json:"total" example:"50"`
	Limit   int            `json:"limit" example:"20"`
	Offset  int            `json:"offset" example:"0"`
}

type StoreListResponse struct {
	Stores []Store `json:"stores"`
	Total  int     `json:"total" example:"12"`
}

type IndianOffersResponse struct {
	Offers []IndianPaymentOffer `json:"offers"`
	Total  int                  `json:"total" example:"15"`
}
