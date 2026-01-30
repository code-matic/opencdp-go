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
	defaultBaseURL               = "https://api.opencdp.io/gateway/data-gateway"
	defaultTimeout               = 10000
	defaultMaxConcurrentRequests = 10
	maxAllowedConcurrentRequests = 30
	version                      = "0.1.0"
)

// Client is the main entry point for the CDP SDK.
type Client struct {
	config     CDPConfig
	httpClient *http.Client
	logger     *Logger
	cio        *CIOIntegration
	semaphore  chan struct{}
}

// NewClient creates a new CDP Client instance.
func NewClient(config CDPConfig) *Client {
	if config.CDPEndpoint == "" {
		config.CDPEndpoint = defaultBaseURL
	}
	if config.Timeout <= 0 {
		config.Timeout = defaultTimeout
	}
	if config.MaxConcurrentRequests <= 0 {
		config.MaxConcurrentRequests = defaultMaxConcurrentRequests
	}
	if config.MaxConcurrentRequests > maxAllowedConcurrentRequests {
		config.MaxConcurrentRequests = maxAllowedConcurrentRequests
	}

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
// GET /v1/health/ping
func (c *Client) Ping(ctx context.Context) error {
	c.logger.Debug("Pinging CDP service")
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.CDPEndpoint+"/v1/health/ping", nil)
	if err != nil {
		return c.handleError(ctx, err, "Failed to create ping request")
	}

	resp, err := c.doRequest(req)
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
func (c *Client) Identify(ctx context.Context, userID string, traits map[string]interface{}) error {
	if err := validateIdentifier(userID); err != nil {
		return c.handleError(ctx, err, "Validation failed for Identify")
	}

	payload := IdentifyPayload{
		UserID: userID,
		Traits: traits,
	}

	c.logger.Debug("Identifying user", "userId", userID)

	// 1. Send to CDP
	if err := c.post(ctx, "/v1/persons/identify", payload); err != nil {
		return c.handleError(ctx, err, "CDP Identify failed")
	}

	// 2. Dual-write to Customer.io (Best effort)
	if c.config.SendToCustomerIO {
		if err := c.cio.Identify(userID, traits); err != nil {
			c.logger.Warn("Customer.io Identify failed", "error", err)
			// We do not fail the main request if the secondary integration fails, unless strict consistency is required (not implemented here)
		}
	}

	return nil
}

// Track sends an event to the CDP and optionally to Customer.io.
func (c *Client) Track(ctx context.Context, userID string, eventName string, properties map[string]interface{}) error {
	if err := validateIdentifier(userID); err != nil {
		return c.handleError(ctx, err, "Validation failed for Track (userId)")
	}
	if err := validateEventName(eventName); err != nil {
		return c.handleError(ctx, err, "Validation failed for Track (eventName)")
	}
	if err := validateProperties(properties); err != nil {
		return c.handleError(ctx, err, "Validation failed for Track (properties)")
	}

	payload := TrackPayload{
		UserID:     userID,
		EventName:  eventName,
		Properties: properties,
	}

	c.logger.Debug("Tracking event", "userId", userID, "event", eventName)

	// 1. Send to CDP
	if err := c.post(ctx, "/v1/persons/track", payload); err != nil {
		return c.handleError(ctx, err, "CDP Track failed")
	}

	// 2. Dual-write to Customer.io
	if c.config.SendToCustomerIO {
		if err := c.cio.Track(userID, eventName, properties); err != nil {
			c.logger.Warn("Customer.io Track failed", "error", err)
		}
	}

	return nil
}

// RegisterDevice registers a device for push notifications.
func (c *Client) RegisterDevice(ctx context.Context, payload DevicePayload) error {
	if err := validateIdentifier(payload.UserID); err != nil {
		return c.handleError(ctx, err, "Validation failed for RegisterDevice")
	}
	if payload.DeviceID == "" {
		return c.handleError(ctx, NewCDPValidationError("device ID required"), "Validation failed")
	}
	if payload.DeviceToken == "" {
		return c.handleError(ctx, NewCDPValidationError("device token required"), "Validation failed")
	}

	c.logger.Debug("Registering device", "userId", payload.UserID, "deviceId", payload.DeviceID)

	// 1. Dual-write to Customer.io (Best effort) - uses DeviceID as the stable identifier
	if c.config.SendToCustomerIO {
		if err := c.cio.AddDevice(payload.UserID, payload.DeviceID, payload.Platform, payload.Attributes); err != nil {
			c.logger.Warn("Customer.io AddDevice failed", "error", err)
		}
	}

	// 2. Send to CDP
	// Assuming /v1/persons/registerDevice endpoint
	if err := c.post(ctx, "/v1/persons/registerDevice", payload); err != nil {
		return c.handleError(ctx, err, "RegisterDevice failed")
	}

	return nil
}

// Internal helper to perform POST requests
func (c *Client) post(ctx context.Context, path string, body interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.config.CDPEndpoint+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
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

// doRequest executes the HTTP request with concurrency control and headers.
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	// Acquire semaphore
	select {
	case c.semaphore <- struct{}{}:
		defer func() { <-c.semaphore }()
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}

	req.Header.Set("Authorization", "Bearer "+c.config.CDPAPIKey)
	req.Header.Set("User-Agent", fmt.Sprintf("opencdp-go-sdk/%s", version))

	return c.httpClient.Do(req)
}

// handleError processes errors based on FailOnException config.
func (c *Client) handleError(ctx context.Context, err error, msg string) error {
	c.logger.LogError(ctx, err, msg)
	if c.config.FailOnException {
		return NewCDPError("API_ERROR", msg, err)
	}
	return nil
}
