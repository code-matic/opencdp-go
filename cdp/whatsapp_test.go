package cdp_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/codematic/opencdp-go/cdp"
	"github.com/stretchr/testify/assert"
)

func whatsAppPayload() cdp.WhatsAppPayload {
	return cdp.WhatsAppPayload{
		Identifiers:            cdp.Identifiers{ID: "u1"},
		TransactionalMessageID: "WA_1",
	}
}

// countingServer replies with status and counts how many requests it received.
func countingServer(status int, hits *int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(hits, 1)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
}

func TestSendWhatsApp_SendsGatewayShape(t *testing.T) {
	var body map[string]interface{}
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/send/whatsapp", r.URL.Path)
		assert.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := cdp.NewClient(mockConfig(server.URL, cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true}))
	defer client.Close()

	payload := whatsAppPayload()
	payload.Identifiers = cdp.Identifiers{CdpID: "cdp_123"}
	payload.TemplateVariables = &cdp.WhatsAppTemplateVars{Body: map[string]interface{}{"1": "{{trigger.order}}"}}
	payload.MessageData = map[string]interface{}{"order": "ORD-1"}

	assert.NoError(t, client.SendWhatsApp(context.Background(), payload))
	assert.Equal(t, map[string]interface{}{
		"identifiers":              map[string]interface{}{"cdp_id": "cdp_123"},
		"transactional_message_id": "WA_1",
		"template_variables":       map[string]interface{}{"body": map[string]interface{}{"1": "{{trigger.order}}"}},
		"message_data":             map[string]interface{}{"order": "ORD-1"},
	}, body)
}

func TestSendWhatsApp_Validation(t *testing.T) {
	client := cdp.NewClient(mockConfig("http://127.0.0.1:1", cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true}))
	defer client.Close()

	tests := []struct {
		name   string
		mutate func(p *cdp.WhatsAppPayload)
	}{
		{"named slot key", func(p *cdp.WhatsAppPayload) {
			p.TemplateVariables = &cdp.WhatsAppTemplateVars{Body: map[string]interface{}{"name": "Jane"}}
		}},
		{"zero slot key", func(p *cdp.WhatsAppPayload) {
			p.TemplateVariables = &cdp.WhatsAppTemplateVars{Button: map[string]interface{}{"0": "x"}}
		}},
		{"invalid phone", func(p *cdp.WhatsAppPayload) { p.To = "0803" }},
		{"cio_id identifier", func(p *cdp.WhatsAppPayload) { p.Identifiers = cdp.Identifiers{CioID: "c1"} }},
		{"phone identifier", func(p *cdp.WhatsAppPayload) { p.Identifiers = cdp.Identifiers{ID: "u1", Phone: "+14155551234"} }},
		{"missing transactional id", func(p *cdp.WhatsAppPayload) { p.TransactionalMessageID = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := whatsAppPayload()
			tt.mutate(&payload)
			assert.Error(t, client.SendWhatsApp(context.Background(), payload))
		})
	}
}

func TestSendWhatsApp_DoesNotFailOverOnClientError(t *testing.T) {
	var primaryHits, fallbackHits int32
	primary := countingServer(http.StatusBadRequest, &primaryHits)
	defer primary.Close()
	fallback := countingServer(http.StatusOK, &fallbackHits)
	defer fallback.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPAPIKey: "key", CDPEndpoint: primary.URL, CDPFallbackEndpoints: []string{fallback.URL}, FailOnException: true,
	})
	defer client.Close()

	assert.Error(t, client.SendWhatsApp(context.Background(), whatsAppPayload()))
	assert.Equal(t, int32(1), atomic.LoadInt32(&primaryHits))
	assert.Equal(t, int32(0), atomic.LoadInt32(&fallbackHits))
}

func TestSendWhatsApp_FailsOverWhenCloudflareNeverReachedTheGateway(t *testing.T) {
	for _, status := range []int{521, 523, 525, 526} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var primaryHits, fallbackHits int32
			primary := countingServer(status, &primaryHits)
			defer primary.Close()
			fallback := countingServer(http.StatusOK, &fallbackHits)
			defer fallback.Close()

			client := cdp.NewClient(cdp.CDPConfig{
				CDPAPIKey: "key", CDPEndpoint: primary.URL, CDPFallbackEndpoints: []string{fallback.URL}, FailOnException: true,
			})
			defer client.Close()

			assert.NoError(t, client.SendWhatsApp(context.Background(), whatsAppPayload()))
			assert.Equal(t, int32(1), atomic.LoadInt32(&fallbackHits))
		})
	}
}

func TestSendWhatsApp_DoesNotFailOverWhenTheGatewayMayHaveQueuedIt(t *testing.T) {
	for _, status := range []int{500, 502, 503, 504, 520, 522, 524} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			var primaryHits, fallbackHits int32
			primary := countingServer(status, &primaryHits)
			defer primary.Close()
			fallback := countingServer(http.StatusOK, &fallbackHits)
			defer fallback.Close()

			client := cdp.NewClient(cdp.CDPConfig{
				CDPAPIKey: "key", CDPEndpoint: primary.URL, CDPFallbackEndpoints: []string{fallback.URL}, FailOnException: true,
			})
			defer client.Close()

			assert.Error(t, client.SendWhatsApp(context.Background(), whatsAppPayload()))
			assert.Equal(t, int32(1), atomic.LoadInt32(&primaryHits))
			assert.Equal(t, int32(0), atomic.LoadInt32(&fallbackHits))
		})
	}
}

func TestSendWhatsApp_FailsOverWhenPrimaryRefusesConnection(t *testing.T) {
	// Reserve a port, then close it so connecting is refused.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	deadURL := "http://" + listener.Addr().String()
	listener.Close()

	var fallbackHits int32
	fallback := countingServer(http.StatusOK, &fallbackHits)
	defer fallback.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPAPIKey: "key", CDPEndpoint: deadURL, CDPFallbackEndpoints: []string{fallback.URL}, FailOnException: true,
	})
	defer client.Close()

	assert.NoError(t, client.SendWhatsApp(context.Background(), whatsAppPayload()))
	assert.Equal(t, int32(1), atomic.LoadInt32(&fallbackHits))
}

func TestSendWhatsApp_DoesNotFailOverAfterTimeout(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slow.Close()
	var fallbackHits int32
	fallback := countingServer(http.StatusOK, &fallbackHits)
	defer fallback.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPAPIKey: "key", CDPEndpoint: slow.URL, CDPFallbackEndpoints: []string{fallback.URL},
		FailOnException: true, Timeout: 50,
	})
	defer client.Close()

	assert.Error(t, client.SendWhatsApp(context.Background(), whatsAppPayload()))
	assert.Equal(t, int32(0), atomic.LoadInt32(&fallbackHits))
}

func TestSendWhatsApp_DoesNotFollowRedirectOrFailOver(t *testing.T) {
	// The redirect target refuses connections. If the SDK followed the redirect, that dial failure
	// would look like the primary was never reached and the send would be retried on the fallback.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	deadURL := "http://" + listener.Addr().String()
	listener.Close()

	var primaryHits, fallbackHits int32
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&primaryHits, 1)
		http.Redirect(w, r, deadURL+r.URL.Path, http.StatusTemporaryRedirect)
	}))
	defer primary.Close()
	fallback := countingServer(http.StatusOK, &fallbackHits)
	defer fallback.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPAPIKey: "key", CDPEndpoint: primary.URL, CDPFallbackEndpoints: []string{fallback.URL}, FailOnException: true,
	})
	defer client.Close()

	err = client.SendWhatsApp(context.Background(), whatsAppPayload())

	assert.ErrorContains(t, err, "307")
	assert.Equal(t, int32(1), atomic.LoadInt32(&primaryHits))
	assert.Equal(t, int32(0), atomic.LoadInt32(&fallbackHits))
}

func TestIdentify_StillFollowsRedirects(t *testing.T) {
	var targetHits int32
	target := countingServer(http.StatusOK, &targetHits)
	defer target.Close()
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+r.URL.Path, http.StatusTemporaryRedirect)
	}))
	defer primary.Close()

	client := cdp.NewClient(mockConfig(primary.URL, cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true}))
	defer client.Close()

	assert.NoError(t, client.Identify(context.Background(), "u1", map[string]interface{}{"plan": "pro"}))
	assert.Equal(t, int32(1), atomic.LoadInt32(&targetHits))
}
