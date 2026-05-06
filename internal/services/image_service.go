package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ImageService handles image proxy and CDN integration for game cover art
type ImageService struct {
	cacheService *CacheService
	logger       *slog.Logger
}

// NewImageService creates a new image service
func NewImageService(cacheService *CacheService, logger *slog.Logger) *ImageService {
	return &ImageService{
		cacheService: cacheService,
		logger:       logger,
	}
}

// ImageQuality presets
const (
	QualityLow    = 60
	QualityMedium = 80
	QualityHigh   = 95
)

// ImageSizes for responsive loading
var ImageSizes = map[string]struct {
	Width  int
	Height int
}{
	"thumbnail": {150, 200},
	"small":    {300, 400},
	"medium":   {600, 800},
	"large":    {1200, 1600},
}

// GetProxiedImageURL returns a proxied URL for the image
func (s *ImageService) GetProxiedImageURL(originalURL string, size string, quality int) string {
	// In production, this would return a URL to your CDN/proxy service
	// For now, return the original URL with query parameters
	
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return originalURL
	}
	
	query := parsedURL.Query()
	query.Set("size", size)
	query.Set("quality", fmt.Sprintf("%d", quality))
	parsedURL.RawQuery = query.Encode()
	
	return parsedURL.String()
}

// ProxyImage fetches and proxies an image with optional resizing and quality adjustment
func (s *ImageService) ProxyImage(ctx context.Context, originalURL string, size string, quality int) ([]byte, string, error) {
	// Check cache first
	cacheKey := s.getCacheKey(originalURL, size, quality)
	cached, err := s.cacheService.GetOrSetJSON(ctx, cacheKey, 24*time.Hour, func() (interface{}, error) {
		return s.fetchAndOptimizeImage(ctx, originalURL, size, quality)
	})
	
	if err != nil {
		return nil, "", fmt.Errorf("failed to get image from cache: %w", err)
	}
	
	// Determine content type
	contentType := "image/jpeg"
	if strings.HasSuffix(originalURL, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(originalURL, ".webp") {
		contentType = "image/webp"
	}
	
	return cached, contentType, nil
}

// fetchAndOptimizeImage fetches the original image and optimizes it
func (s *ImageService) fetchAndOptimizeImage(ctx context.Context, originalURL string, size string, quality int) ([]byte, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	// Fetch the image
	resp, err := client.Get(originalURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	
	// Read image data
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}
	
	// TODO: Implement image optimization (resizing, quality adjustment)
	// This would require an image processing library like imaging or disintegration/imaging
	// For now, return the original data
	s.logger.Debug("fetched image (optimization not implemented)", "url", originalURL, "size", size, "quality", quality)
	
	return data, nil
}

// getCacheKey generates a cache key for the image
func (s *ImageService) getCacheKey(originalURL string, size string, quality int) string {
	return fmt.Sprintf("image:%s:%s:%d", url.QueryEscape(originalURL), size, quality)
}

// GetOptimizedImageURL returns an optimized CDN URL for the image
// In production, this would integrate with a CDN like Cloudflare, Cloudinary, or imgix
func (s *ImageService) GetOptimizedImageURL(originalURL string, size string) string {
	// Example CDN integration patterns:
	// Cloudinary: https://res.cloudinary.com/{cloud_name}/image/fetch/w_{width},h_{height},q_{quality}/{url}
	// imgix: https://{domain}.imgix.net/{path}?w={width}&h={height}&q={quality}&auto=format
	
	// For now, return the original URL
	// In production, implement actual CDN integration
	return originalURL
}

// CDNProviders represents supported CDN providers
type CDNProvider string

const (
	CDNProviderNone      CDNProvider = "none"
	CDNProviderCloudinary CDNProvider = "cloudinary"
	CDNProviderImgix     CDNProvider = "imgix"
	CDNProviderCloudflare CDNProvider = "cloudflare"
)

// CDNConfig holds CDN configuration
type CDNConfig struct {
	Provider CDNProvider
	CloudName string // For Cloudinary
	Domain    string // For imgix
}

// SetCDNConfig configures the CDN provider
func (s *ImageService) SetCDNConfig(config CDNConfig) {
	// TODO: Implement CDN configuration
	s.logger.Info("CDN configuration would be set", "provider", config.Provider)
}

// InvalidateImageCache invalidates cached images for a specific URL
func (s *ImageService) InvalidateImageCache(ctx context.Context, originalURL string) error {
	// Invalidate all cached versions of this image
	sizes := []string{"thumbnail", "small", "medium", "large"}
	qualities := []int{QualityLow, QualityMedium, QualityHigh}
	
	for _, size := range sizes {
		for _, quality := range qualities {
			cacheKey := s.getCacheKey(originalURL, size, quality)
			if s.cacheService.client != nil {
				if err := s.cacheService.client.Del(ctx, cacheKey).Err(); err != nil {
					s.logger.Warn("failed to invalidate image cache", "key", cacheKey, "error", err)
				}
			}
		}
	}
	
	s.logger.Debug("invalidated image cache", "url", originalURL)
	return nil
}
