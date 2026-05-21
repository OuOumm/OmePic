package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultCloudflareAPIBaseURL = "https://api.cloudflare.com/client/v4"

// ImageURLCachePurger purges one or more public image URLs from an edge cache.
type ImageURLCachePurger interface {
	PurgeURL(ctx context.Context, rawURL string) error
	PurgeURLs(ctx context.Context, rawURLs []string) error
}

type CloudflareCachePurger struct {
	zoneID     string
	apiToken   string
	apiBaseURL string
	client     *http.Client
}

type cloudflarePurgeCacheRequest struct {
	Files []string `json:"files"`
}

type cloudflarePurgeCacheResponse struct {
	Success bool `json:"success"`
}

func NewCloudflareCachePurger(zoneID string, apiToken string, apiBaseURL string, client *http.Client) *CloudflareCachePurger {
	baseURL := strings.TrimRight(strings.TrimSpace(apiBaseURL), "/")
	if baseURL == "" {
		baseURL = defaultCloudflareAPIBaseURL
	}
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &CloudflareCachePurger{
		zoneID:     strings.TrimSpace(zoneID),
		apiToken:   strings.TrimSpace(apiToken),
		apiBaseURL: baseURL,
		client:     client,
	}
}

func (p *CloudflareCachePurger) Configured() bool {
	return p != nil && strings.TrimSpace(p.zoneID) != "" && strings.TrimSpace(p.apiToken) != ""
}

func (p *CloudflareCachePurger) PurgeURL(ctx context.Context, rawURL string) error {
	return p.PurgeURLs(ctx, []string{rawURL})
}

func (p *CloudflareCachePurger) PurgeURLs(ctx context.Context, rawURLs []string) error {
	imageURLs, err := normalizeCloudflarePurgeURLs(rawURLs)
	if err != nil {
		return err
	}
	if !p.Configured() {
		return fmt.Errorf("%w: cloudflare zone id or api token is not configured", ErrDependencyUnavailable)
	}

	endpoint := strings.TrimRight(p.apiBaseURL, "/") + "/zones/" + url.PathEscape(p.zoneID) + "/purge_cache"
	payload := cloudflarePurgeCacheRequest{Files: imageURLs}
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(payload); err != nil {
		return fmt.Errorf("%w: cloudflare purge payload encode failed", ErrDependencyUnavailable)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, &body)
	if err != nil {
		return fmt.Errorf("%w: cloudflare purge request build failed", ErrDependencyUnavailable)
	}
	request.Header.Set("Authorization", "Bearer "+p.apiToken)
	request.Header.Set("Content-Type", "application/json")

	response, err := p.client.Do(request)
	if err != nil {
		return fmt.Errorf("%w: cloudflare purge request failed", ErrDependencyUnavailable)
	}
	defer response.Body.Close()

	responseBody, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("%w: cloudflare purge returned status %d", ErrDependencyUnavailable, response.StatusCode)
	}
	if readErr != nil {
		return fmt.Errorf("%w: cloudflare purge response read failed", ErrDependencyUnavailable)
	}

	var decoded cloudflarePurgeCacheResponse
	if err := json.Unmarshal(responseBody, &decoded); err != nil {
		return fmt.Errorf("%w: cloudflare purge response decode failed", ErrDependencyUnavailable)
	}
	if !decoded.Success {
		return fmt.Errorf("%w: cloudflare purge was not accepted", ErrDependencyUnavailable)
	}
	return nil
}

func normalizeCloudflarePurgeURL(rawURL string) (string, error) {
	trimmed := strings.TrimSpace(rawURL)
	if trimmed == "" {
		return "", WithUserMessage(ErrInvalidInput, "cloudflare purge url is required")
	}
	parsed, err := url.Parse(trimmed)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", WithUserMessage(ErrInvalidInput, "cloudflare purge url must be an http or https URL")
	}
	parsed.Fragment = ""
	return parsed.String(), nil
}

func normalizeCloudflarePurgeURLs(rawURLs []string) ([]string, error) {
	if len(rawURLs) == 0 {
		return nil, WithUserMessage(ErrInvalidInput, "cloudflare purge urls are required")
	}
	imageURLs := make([]string, 0, len(rawURLs))
	for _, rawURL := range rawURLs {
		imageURL, err := normalizeCloudflarePurgeURL(rawURL)
		if err != nil {
			return nil, err
		}
		imageURLs = append(imageURLs, imageURL)
	}
	return imageURLs, nil
}
