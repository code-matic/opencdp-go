package cdp_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/codematic/opencdp-go/cdp"
	"github.com/stretchr/testify/assert"
)

func setupMockServerForValidation(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func defaultHandlerForValidation(t *testing.T, expectedPath string, expectedMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, expectedPath, r.URL.Path)
		assert.Equal(t, expectedMethod, r.Method)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}
}

// Test BCC validation
func TestSendEmail_BCC_Validation(t *testing.T) {
	server := setupMockServerForValidation(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// Valid BCC
	err := client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
		BCC:                    []string{"bcc1@example.com", "bcc2@example.com"},
	})
	assert.NoError(t, err)

	// Invalid BCC email
	err = client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
		BCC:                    []string{"invalid-email"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid bcc email")
}

// Test CC validation
func TestSendEmail_CC_Validation(t *testing.T) {
	server := setupMockServerForValidation(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// Invalid CC email
	err := client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
		CC:                     []string{"valid@example.com", "not-valid"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cc email")
}

// Test template vs raw email validation
func TestSendEmail_TemplateVsRaw(t *testing.T) {
	server := setupMockServerForValidation(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// Template email - should work
	err := client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
	})
	assert.NoError(t, err)

	// Raw email without required fields - should fail
	err = client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		Body: "Hello",
		// Missing subject and from
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "subject is required")

	// Raw email with all required fields - should work
	err = client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		Body:    "Hello",
		Subject: "Test",
		From:    "sender@example.com",
	})
	assert.NoError(t, err)
}

// Test identifier "exactly one" validation
func TestSendEmail_IdentifierValidation(t *testing.T) {
	server := setupMockServerForValidation(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// No identifiers - should fail
	err := client.SendEmail(context.Background(), cdp.EmailPayload{
		To:                     "test@example.com",
		Identifiers:            cdp.Identifiers{},
		TransactionalMessageID: "WELCOME",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one")

	// Multiple identifiers - should fail
	err = client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID:    "user1",
			Email: "user@example.com",
		},
		TransactionalMessageID: "WELCOME",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exactly one")

	// Exactly one identifier - should work
	err = client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
	})
	assert.NoError(t, err)
}

// Test from/reply_to validation
func TestSendEmail_FromReplyToValidation(t *testing.T) {
	server := setupMockServerForValidation(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// Invalid from email
	err := client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
		From:                   "invalid-email",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")

	// Invalid reply_to email
	err = client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
		ReplyTo:                "not-an-email",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid email format")
}

// Test send_at validation
func TestSendEmail_SendAtValidation(t *testing.T) {
	server := setupMockServerForValidation(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// Negative send_at - should fail
	err := client.SendEmail(context.Background(), cdp.EmailPayload{
		To: "test@example.com",
		Identifiers: cdp.Identifiers{
			ID: "user1",
		},
		TransactionalMessageID: "WELCOME",
		SendAt:                 -100,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a positive integer")
}

// Test push identifier validation
func TestSendPush_IdentifierValidation(t *testing.T) {
	server := setupMockServerForValidation(t, defaultHandlerForValidation(t, "/v1/send/push", "POST"))
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// No identifiers - should fail
	err := client.SendPush(context.Background(), cdp.PushPayload{
		Identifiers:            cdp.Identifiers{},
		TransactionalMessageID: "PUSH1",
	})
	assert.Error(t, err)

	// Multiple identifiers - should fail
	err = client.SendPush(context.Background(), cdp.PushPayload{
		Identifiers: cdp.Identifiers{
			ID:    "user1",
			Email: "user@example.com",
		},
		TransactionalMessageID: "PUSH1",
	})
	assert.Error(t, err)
}

// Test SMS identifier validation
func TestSendSms_IdentifierValidation(t *testing.T) {
	server := setupMockServerForValidation(t, defaultHandlerForValidation(t, "/v1/send/sms", "POST"))
	defer server.Close()

	client := cdp.NewClient(cdp.CDPConfig{CDPEndpoint: server.URL, CDPAPIKey: "key", FailOnException: true})
	defer client.Close()

	// No identifiers - should fail
	err := client.SendSms(context.Background(), cdp.SmsPayload{
		Identifiers:            cdp.Identifiers{},
		TransactionalMessageID: "SMS1",
	})
	assert.Error(t, err)
}
