package cdp

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// CDPConfig holds configuration for the CDP client.
type CDPConfig struct {
	// API Key for the CDP
	CDPAPIKey string
	// Optional custom endpoint (defaults to production)
	CDPEndpoint string
	// Optional fallback gateway base URLs
	CDPFallbackEndpoints []string

	// Request timeout in milliseconds (default: 10000)
	Timeout int
	// If true, methods will return errors. If false, errors are logged and nil is returned.
	FailOnException bool
	// Maximum number of concurrent requests (default: 10, max: 30)
	MaxConcurrentRequests int
	// Enable debug logging
	Debug bool
	// Custom logger (optional)
	Logger *slog.Logger

	// Dual-write configuration
	SendToCustomerIO bool
	CustomerIO       *CustomerIOConfig
}

// CustomerIOConfig holds configuration for the optional Customer.io integration.
type CustomerIOConfig struct {
	SiteID string
	APIKey string
	Region string // "us" or "eu"
}

// Identifiers represents the user identifiers.
type Identifiers struct {
	ID    string `json:"id,omitempty"`
	Email string `json:"email,omitempty"`
	CdpID string `json:"cdp_id,omitempty"`
	// Deprecated: the gateway does not accept phone as an identifier; requests that set it are rejected.
	Phone string `json:"phone,omitempty"`
	// Deprecated: the gateway does not accept cio_id; use CdpID.
	CioID string `json:"cio_id,omitempty"`
}

// IdentifyPayload represents the data for an identify call.
type IdentifyPayload struct {
	Identifier string                 `json:"identifier"`
	Properties map[string]interface{} `json:"properties"`
}

// TrackPayload represents the data for a track call.
type TrackPayload struct {
	Identifier string                 `json:"identifier"`
	EventName  string                 `json:"eventName"`
	Properties map[string]interface{} `json:"properties"`
}

// EmailPayload represents the data for sending an email.
type EmailPayload struct {
	To                      string                 `json:"to"`
	Identifiers             Identifiers            `json:"identifiers"`
	TransactionalMessageID  string                 `json:"transactional_message_id,omitempty"`
	Subject                 string                 `json:"subject,omitempty"`
	Body                    string                 `json:"body,omitempty"`
	BodyPlain               string                 `json:"body_plain,omitempty"`
	From                    string                 `json:"from,omitempty"`
	ReplyTo                 string                 `json:"reply_to,omitempty"`
	BCC                     []string               `json:"bcc,omitempty"`
	CC                      []string               `json:"cc,omitempty"`
	Preheader               string                 `json:"preheader,omitempty"`
	AmpBody                 string                 `json:"amp_body,omitempty"`
	PlaintextBody           string                 `json:"plaintext_body,omitempty"`
	Headers                 map[string]interface{} `json:"headers,omitempty"`
	MessageData             map[string]interface{} `json:"message_data,omitempty"`
	FakeBCC                 bool                   `json:"fake_bcc,omitempty"`
	DisableMessageRetention bool                   `json:"disable_message_retention,omitempty"`
	SendToUnsubscribed      bool                   `json:"send_to_unsubscribed,omitempty"`
	QueueDraft              bool                   `json:"queue_draft,omitempty"`
	DisableCSSPreprocessing bool                   `json:"disable_css_preprocessing,omitempty"`
	Language                string                 `json:"language,omitempty"`
	// Attachments maps filename to base64 content. Max 5 files, 2 MB decoded in total.
	// Use Attach or AttachFile to add files without encoding them yourself.
	Attachments map[string]string `json:"attachments,omitempty"`
	// Unsupported fields included for compatibility, will trigger warnings
	SendAt  int64 `json:"send_at,omitempty"`
	Tracked bool  `json:"tracked,omitempty"`
}

// PushPayload represents the data for sending a push notification.
type PushPayload struct {
	Identifiers            Identifiers            `json:"identifiers"`
	TransactionalMessageID string                 `json:"transactional_message_id"`
	Title                  string                 `json:"title,omitempty"`
	Body                   string                 `json:"body,omitempty"`
	MessageData            map[string]interface{} `json:"message_data,omitempty"`
}

// SmsPayload represents the data for sending an SMS.
type SmsPayload struct {
	Identifiers            Identifiers            `json:"identifiers"`
	To                     string                 `json:"to,omitempty"`
	From                   string                 `json:"from,omitempty"`
	TransactionalMessageID string                 `json:"transactional_message_id,omitempty"`
	Body                   string                 `json:"body,omitempty"`
	MessageData            map[string]interface{} `json:"message_data,omitempty"`
}

// WhatsAppPayload represents the data for sending a WhatsApp message.
type WhatsAppPayload struct {
	Identifiers            Identifiers            `json:"identifiers"`
	TransactionalMessageID string                 `json:"transactional_message_id"`
	To                     string                 `json:"to,omitempty"`
	TemplateVariables      *WhatsAppTemplateVars  `json:"template_variables,omitempty"`
	MessageData            map[string]interface{} `json:"message_data,omitempty"`
}

// WhatsAppTemplateVars sets template slots by section. Keys are the template's slot numbers
// ("1", "2", ...) and values may use Liquid, e.g. "{{trigger.order_number}}".
// When set, it replaces all variables saved on the transactional, so include every section the
// template needs.
type WhatsAppTemplateVars struct {
	Header map[string]interface{} `json:"header,omitempty"`
	Body   map[string]interface{} `json:"body,omitempty"`
	Button map[string]interface{} `json:"button,omitempty"`
}

// DevicePayload represents data for registering a device.
type DevicePayload struct {
	Identifier   string                 `json:"identifier"`
	DeviceID     string                 `json:"deviceId"`
	Platform     string                 `json:"platform"`
	FcmToken     string                 `json:"fcmToken,omitempty"`
	ApnToken     string                 `json:"apnToken,omitempty"`
	Name         string                 `json:"name,omitempty"`
	OSVersion    string                 `json:"osVersion,omitempty"`
	Model        string                 `json:"model,omitempty"`
	AppVersion   string                 `json:"appVersion,omitempty"`
	LastActiveAt string                 `json:"last_active_at,omitempty"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

// Attach adds a file to the email, base64-encoding data. To add content that is already
// base64, set it on the Attachments map directly.
func (p *EmailPayload) Attach(filename string, data []byte) {
	if p.Attachments == nil {
		p.Attachments = map[string]string{}
	}
	p.Attachments[filename] = base64.StdEncoding.EncodeToString(data)
}

// AttachFile reads the file at path and attaches it under its base name.
func (p *EmailPayload) AttachFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read attachment %q: %w", path, err)
	}
	p.Attach(filepath.Base(path), data)
	return nil
}
