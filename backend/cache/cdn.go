package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// CDNCache handles CDN integration for static assets
type CDNCache struct {
	// CDN configuration
	cdnURL     string
	originURL  string
	purgeURL   string
	apiKey     string

	// Local cache for CDN URLs
	urlCache   map[string]string
	mu         sync.RWMutex

	// Asset versioning
	versionMap map[string]string

	// HTTP client
	client *http.Client
}

// CDNConfig holds CDN configuration
type CDNConfig struct {
	CDNURL    string
	OriginURL string
	PurgeURL  string
	APIKey    string
	Timeout   time.Duration
}

// NewCDNCache creates a new CDN cache
func NewCDNCache(config *CDNConfig) *CDNCache {
	return &CDNCache{
		cdnURL:     config.CDNURL,
		originURL:  config.OriginURL,
		purgeURL:   config.PurgeURL,
		apiKey:     config.APIKey,
		urlCache:   make(map[string]string),
		versionMap: make(map[string]string),
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// GetAssetURL returns the CDN URL for an asset
func (c *CDNCache) GetAssetURL(assetPath string) string {
	c.mu.RLock()
	if url, ok := c.urlCache[assetPath]; ok {
		c.mu.RUnlock()
		return url
	}
	c.mu.RUnlock()

	// Generate versioned URL
	version := c.getAssetVersion(assetPath)
	url := fmt.Sprintf("%s/%s?v=%s", c.cdnURL, assetPath, version)

	c.mu.Lock()
	c.urlCache[assetPath] = url
	c.mu.Unlock()

	return url
}

// PurgeAsset purges an asset from CDN
func (c *CDNCache) PurgeAsset(ctx context.Context, assetPath string) error {
	if c.purgeURL == "" {
		return fmt.Errorf("purge URL not configured")
	}

	url := fmt.Sprintf("%s/purge/%s", c.purgeURL, assetPath)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CDN purge failed with status %d", resp.StatusCode)
	}

	// Clear local cache
	c.mu.Lock()
	delete(c.urlCache, assetPath)
	delete(c.versionMap, assetPath)
	c.mu.Unlock()

	return nil
}

// PurgeAll purges all assets from CDN
func (c *CDNCache) PurgeAll(ctx context.Context) error {
	if c.purgeURL == "" {
		return fmt.Errorf("purge URL not configured")
	}

	url := fmt.Sprintf("%s/purge-all", c.purgeURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CDN purge all failed with status %d", resp.StatusCode)
	}

	// Clear all local caches
	c.mu.Lock()
	c.urlCache = make(map[string]string)
	c.versionMap = make(map[string]string)
	c.mu.Unlock()

	return nil
}

// SetAssetVersion manually sets asset version (for cache busting)
func (c *CDNCache) SetAssetVersion(assetPath, version string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.versionMap[assetPath] = version
	delete(c.urlCache, assetPath) // Clear cached URL
}

// getAssetVersion gets or generates asset version hash
func (c *CDNCache) getAssetVersion(assetPath string) string {
	c.mu.RLock()
	if version, ok := c.versionMap[assetPath]; ok {
		c.mu.RUnlock()
		return version
	}
	c.mu.RUnlock()

	// Generate hash from current timestamp (or file content hash in production)
	hash := sha256.Sum256([]byte(assetPath + time.Now().Format("20060102")))
	version := hex.EncodeToString(hash[:8])

	c.mu.Lock()
	c.versionMap[assetPath] = version
	c.mu.Unlock()

	return version
}

// DownloadAsset downloads an asset from CDN (with local caching)
func (c *CDNCache) DownloadAsset(ctx context.Context, assetPath string) ([]byte, error) {
	url := c.GetAssetURL(assetPath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("asset download failed with status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// SetCacheControl sets cache control headers for HTTP responses
func SetCacheControl(w http.ResponseWriter, maxAge time.Duration, public bool) {
	cacheType := "private"
	if public {
		cacheType = "public"
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("%s, max-age=%d", cacheType, int(maxAge.Seconds())))

	// Set Expires header
	expires := time.Now().Add(maxAge)
	w.Header().Set("Expires", expires.Format(http.TimeFormat))
}

// SetNoCache disables caching for HTTP responses
func SetNoCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

// ETagMiddleware adds ETag support for HTTP responses
func ETagMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture response
		rw := &responseWriter{
			ResponseWriter: w,
			body:          make([]byte, 0),
		}

		next.ServeHTTP(rw, r)

		// Generate ETag
		hash := sha256.Sum256(rw.body)
		etag := hex.EncodeToString(hash[:])

		// Check If-None-Match
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		// Set ETag header
		w.Header().Set("ETag", etag)
		w.Write(rw.body)
	})
}

// responseWriter captures response body for ETag generation
type responseWriter struct {
	http.ResponseWriter
	body []byte
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return len(b), nil
}
