package cdp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultTimeout               = 10000
	defaultMaxConcurrentRequests = 10
	maxAllowedConcurrentRequests = 30
	version                      = "0.2.0"
)

// Client is the main entry point for the CDP SDK.
type Client struct {
	config     CDPConfig
	baseURLs   []string
	httpClient *http.Client
	logger     *Logger
	cio        *CIOIntegration
	semaphore  chan struct{}
}

// NewClient creates a new CDP Client instance.
func NewClient(config CDPConfig) *Client {
	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}
	if config.MaxConcurrentRequests <= 0 {
		config.MaxConcurrentRequests = defaultMaxConcurrentRequests
	}
	if config.MaxConcurrentRequests > maxAllowedConcurrentRequests {
		config.MaxConcurrentRequests = maxAllowedConcurrentRequests
	}

	baseURLs := ResolveAllBaseURLs(config.CDPEndpoint, config.CDPFallbackEndpoints)
	logger := NewLogger(config.Logger, config.Debug)

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   time.Duration(config.Timeout) * time.Millisecond,
		Transport: transport,
	}

	return &Client{
		config:     config,
		baseURLs:   baseURLs,
		httpClient: httpClient,
		logger:     logger,
		cio:        NewCIOIntegration(&config),
		semaphore:  make(chan struct{}, config.MaxConcurrentRequests),
	}
}

// Close cleans up resources used by the client.
func (c *Client) Close() {
	c.logger.Debug("Closing CDP client resources")
	if t, ok := c.httpClient.Transport.(*http.Transport); ok {
		t.CloseIdleConnections()
	}
	close(c.semaphore)
}

// Ping checks the health of the CDP connection.
func (c *Client) Ping(ctx context.Context) error {
	c.logger.Debug("Pinging CDP service")
	resp, err := c.requestWithFailover(ctx, "GET", "/v1/health/ping", nil)
	if err != nil {
		return c.handleError(ctx, err, "Ping request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return c.handleError(ctx, fmt.Errorf("unexpected status code: %d", resp.StatusCode), "Ping failed")
	}
	return nil
}

// Identify sends user identification data to the CDP and optionally to Customer.io.
func (c *Client) Identify(ctx context.Context, identifier string, properties map[string]interface{}) error {
	if err := validateIdentifier(identifier); err != nil {
		return c.handleError(ctx, err, "Validation failed for Identify")
	}

	payload := IdentifyPayload{
		Identifier: identifier,
		Properties: properties,
	}

	c.logger.Debug("Identifying user", "identifier", identifier)

	if err := c.post(ctx, "/v1/persons/identify", payload); err != nil {
		return c.handleError(ctx, err, "CDP Identify failed")
	}

	if c.config.SendToCustomerIO {
		if err := c.cio.Identify(identifier, properties); err != nil {
			c.logger.Warn("Customer.io Identify failed", "error", err)
		}
	}

	return nil
}

// Track sends an event to the CDP and optionally to Customer.io.
func (c *Client) Track(ctx context.Context, identifier string, eventName string, properties map[string]interface{}) error {
	if err := validateIdentifier(identifier); err != nil {
		return c.handleError(ctx, err, "Validation failed for Track (identifier)")
	}
	if err := validateEventName(eventName); err != nil {
		return c.handleError(ctx, err, "Validation failed for Track (eventName)")
	}
	if err := validateProperties(properties); err != nil {
		return c.handleError(ctx, err, "Validation failed for Track (properties)")
	}

	payload := TrackPayload{
		Identifier: identifier,
		EventName:  eventName,
		Properties: properties,
	}

	c.logger.Debug("Tracking event", "identifier", identifier, "eventName", eventName)

	if err := c.post(ctx, "/v1/persons/track", payload); err != nil {
		return c.handleError(ctx, err, "CDP Track failed")
	}

	if c.config.SendToCustomerIO {
		if err := c.cio.Track(identifier, eventName, properties); err != nil {
			c.logger.Warn("Customer.io Track failed", "error", err)
		}
	}

	return nil
}

// RegisterDevice registers a device for push notifications.
func (c *Client) RegisterDevice(ctx context.Context, payload DevicePayload) error {
	if err := validateIdentifier(payload.Identifier); err != nil {
		return c.handleError(ctx, err, "Validation failed for RegisterDevice")
	}
	if payload.DeviceID == "" {
		return c.handleError(ctx, NewCDPValidationError("device ID required"), "Validation failed")
	}
	if payload.FcmToken == "" && payload.ApnToken == "" {
		return c.handleError(ctx, NewCDPValidationError("fcmToken or apnToken required"), "Validation failed")
	}

	c.logger.Debug("Registering device", "identifier", payload.Identifier, "deviceId", payload.DeviceID)

	if c.config.SendToCustomerIO {
		if err := c.cio.AddDevice(payload.Identifier, payload.DeviceID, payload.Platform, payload.Attributes); err != nil {
			c.logger.Warn("Customer.io AddDevice failed", "error", err)
		}
	}

	if err := c.post(ctx, "/v1/persons/registerDevice", payload); err != nil {
		return c.handleError(ctx, err, "RegisterDevice failed")
	}

	return nil
}

func (c *Client) post(ctx context.Context, path string, body interface{}) error {
	resp, err := c.requestWithFailover(ctx, "POST", path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

func (c *Client) requestWithFailover(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var jsonBody []byte
	var err error
	if body != nil {
		jsonBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
	}

	var lastErr error
	for _, baseURL := range c.baseURLs {
		var reader io.Reader
		if jsonBody != nil {
			reader = bytes.NewBuffer(jsonBody)
		}
		req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		if jsonBody != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := c.doRequest(req)
		if err != nil {
			lastErr = err
			if c.config.Debug {
				c.logger.Debug("Gateway unreachable", "baseURL", baseURL, "error", err)
			}
			continue
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		lastErr = fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
		if c.config.Debug {
			c.logger.Debug("Gateway returned error, trying next host", "baseURL", baseURL, "status", resp.StatusCode)
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no gateway hosts configured")
	}
	return nil, lastErr
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}

	req.Header.Set("Authorization", c.config.CDPAPIKey)
	req.Header.Set("User-Agent", fmt.Sprintf("opencdp-go-sdk/%s", version))

	return c.httpClient.Do(req)
}

func (c *Client) handleError(ctx context.Context, err error, msg string) error {
	c.logger.LogError(ctx, err, msg)
	if c.config.FailOnException {
		return NewCDPError("API_ERROR", msg, err)
	}
	return nil
}
