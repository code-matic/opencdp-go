package cdp_test

import (
	"context"
	"encoding/json"
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

func TestSendWhatsApp_FailsOverOn503(t *testing.T) {
	var primaryHits, fallbackHits int32
	primary := countingServer(http.StatusServiceUnavailable, &primaryHits)
	defer primary.Close()
	fallback := countingServer(http.StatusOK, &fallbackHits)
	defer fallback.Close()

	client := cdp.NewClient(cdp.CDPConfig{
		CDPAPIKey: "key", CDPEndpoint: primary.URL, CDPFallbackEndpoints: []string{fallback.URL}, FailOnException: true,
	})
	defer client.Close()

	assert.NoError(t, client.SendWhatsApp(context.Background(), whatsAppPayload()))
	assert.Equal(t, int32(1), atomic.LoadInt32(&fallbackHits))
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
