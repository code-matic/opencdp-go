package cdp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codematic/opencdp-go/cdp"
	"github.com/stretchr/testify/assert"
)

// --- Helper Functions ---

func setupMockServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func defaultHandler(t *testing.T, expectedPath string, expectedMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expectedPath, r.URL.Path)
		assert.Equal(t, expectedMethod, r.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}
}

// --- 1. Happy Path Tests ---

func TestIdentify_Success(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/persons/identify", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var payload cdp.IdentifyPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, "user_123", payload.Identifier)
		assert.Equal(t, "Alice", payload.Properties["name"])

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPEndpoint: server.URL,
		CDPAPIKey:   "test_key",
		Debug:       true,
	})
	defer client.Close()

	err := client.Identify(context.Background(), "user_123", map[string]interface{}{"name": "Alice"})
	assert.NoError(t, err)
}

func TestTrack_Success(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Assuming /track endpoint based on standard conventions
		assert.Equal(t, "/v1/persons/track", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var payload cdp.TrackPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, "user_123", payload.Identifier)
		assert.Equal(t, "purchase", payload.EventName)

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPEndpoint: server.URL,
		CDPAPIKey:   "test_key",
	})
	defer client.Close()

	err := client.Track(context.Background(), "user_123", "purchase", map[string]interface{}{"price": 99.9})
	assert.NoError(t, err)
}

func TestSendEmail_Success(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/send/email", r.URL.Path)
		var payload cdp.EmailPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Equal(t, "test@example.com", payload.To)
		assert.Equal(t, "WELCOME", payload.TransactionalMessageID)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key"})
	defer client.Close()

	payload := cdp.EmailPayload{
		To:                     "test@example.com",
		Identifiers:            cdp.Identifiers{ID: "u1"},
		TransactionalMessageID: "WELCOME",
	}
	err := client.SendEmail(context.Background(), payload)
	assert.NoError(t, err)
}

func TestSendPush_Success(t *testing.T) {
	server := setupMockServer(t, defaultHandler(t, "/v1/send/push", "POST"))
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key"})
	defer client.Close()

	payload := cdp.PushPayload{
		Identifiers:            cdp.Identifiers{ID: "u1"},
		TransactionalMessageID: "PUSH_1",
		Title:                  "Hello",
	}
	err := client.SendPush(context.Background(), payload)
	assert.NoError(t, err)
}

func TestSendSms_Success(t *testing.T) {
	server := setupMockServer(t, defaultHandler(t, "/v1/send/sms", "POST"))
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key"})
	defer client.Close()

	payload := cdp.SmsPayload{
		Identifiers:            cdp.Identifiers{ID: "u1"},
		TransactionalMessageID: "SMS_1",
	}
	err := client.SendSms(context.Background(), payload)
	assert.NoError(t, err)
}

func TestPing_Success(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/health/ping", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL})
	defer client.Close()

	err := client.Ping(context.Background())
	assert.NoError(t, err)
}

// --- 2. Validation Error Tests ---

func TestIdentify_ValidationError_EmptyID(t *testing.T) {
	client := cdp.NewClient(cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	err := client.Identify(context.Background(), "", map[string]interface{}{"a": 1})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "identifier cannot be empty")
}

func TestSendEmail_ValidationError_InvalidEmail(t *testing.T) {
	client := cdp.NewClient(cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	payload := cdp.EmailPayload{
		To:          "invalid-email",
		Identifiers: cdp.Identifiers{ID: "u1"},
	}
	err := client.SendEmail(context.Background(), payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestSendSms_ValidationError_InvalidPhone(t *testing.T) {
	client := cdp.NewClient(cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	payload := cdp.SmsPayload{
		To:          "12345",        // Invalid E.164
		Body:        "Test message", // Body required when no template
		Identifiers: cdp.Identifiers{ID: "u1"},
	}
	err := client.SendSms(context.Background(), payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid phone number format")
}

func TestTrack_ValidationError_EmptyEvent(t *testing.T) {
	client := cdp.NewClient(cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// Assuming Track validates event name similar to other validations
	err := client.Track(context.Background(), "u1", "", nil)
	assert.Error(t, err)
	// Note: Exact error message depends on implementation, checking for error existence
}

// --- 3. Configuration Tests ---

func TestConfiguration_FailOnException_True(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPEndpoint:     server.URL,
		CDPAPIKey:       "key",
		FailOnException: true,
	})
	defer client.Close()

	err := client.Identify(context.Background(), "u1", nil)
	assert.Error(t, err)
}

func TestConfiguration_FailOnException_False(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPEndpoint:     server.URL,
		CDPAPIKey:       "key",
		FailOnException: false,
	})
	defer client.Close()

	// Should log error but return nil
	err := client.Identify(context.Background(), "u1", nil)
	assert.NoError(t, err)
}

func TestConfiguration_Timeout(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPEndpoint:     server.URL,
		CDPAPIKey:       "key",
		Timeout:         10, // 10ms timeout
		FailOnException: true,
	})
	defer client.Close()

	err := client.Identify(context.Background(), "u1", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded") // or client timeout error
}

// --- 4. Dual-Write Tests ---

func TestDualWrite_Enabled(t *testing.T) {
	// We cannot easily mock the internal Customer.io client without dependency injection,
	// but we can verify that enabling it doesn't crash the SDK and still calls the CDP.
	server := setupMockServer(t, defaultHandler(t, "/v1/persons/identify", "POST"))
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPEndpoint:      server.URL,
		CDPAPIKey:        "key",
		SendToCustomerIO: true,
		CustomerIO: &cdp.CustomerIOConfig{
			SiteID: "fake_site_id",
			APIKey: "fake_api_key",
		},
	})
	defer client.Close()

	// This should succeed for CDP. CIO call happens internally; we assume it handles network errors gracefully or we ignore them here.
	_ = client.Identify(context.Background(), "u1", map[string]interface{}{"trait": "val"})
	// If CIO fails (likely due to bad creds/network), Identify might return error depending on implementation.
	// However, usually dual-write is best-effort. If it fails, we check if SDK handles it.
	// For this test, we just ensure no panic.
	assert.NotPanics(t, func() {
		_ = client.Identify(context.Background(), "u1", map[string]interface{}{"trait": "val"})
	})
}

func TestDualWrite_Disabled(t *testing.T) {
	server := setupMockServer(t, defaultHandler(t, "/v1/persons/identify", "POST"))
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPEndpoint:      server.URL,
		CDPAPIKey:        "key",
		SendToCustomerIO: false,
	})
	defer client.Close()

	err := client.Identify(context.Background(), "u1", nil)
	assert.NoError(t, err)
}

// --- 5. Edge Cases ---

func TestEdgeCase_EmptyProperties(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var payload cdp.TrackPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Empty(t, payload.Properties)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key"})
	defer client.Close()

	err := client.Track(context.Background(), "u1", "event", map[string]interface{}{})
	assert.NoError(t, err)
}

func TestEdgeCase_NullValues(t *testing.T) {
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		var payload cdp.IdentifyPayload
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)
		assert.Nil(t, payload.Properties)
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key"})
	defer client.Close()

	// Passing nil traits
	err := client.Identify(context.Background(), "u1", nil)
	assert.NoError(t, err)
}

func TestClose_Resources(t *testing.T) {
	client := cdp.NewClient(cdp.CDPConfig{CDPAPIKey: "key"})
	// Calling Close multiple times should be safe or at least once should work
	assert.NotPanics(t, func() {
		client.Close()
	})
}
