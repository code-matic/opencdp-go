package cdp

import "log/slog"

// CDPConfig holds configuration for the CDP client.
type CDPConfig struct {
	// API Key for the CDP
	CDPAPIKey string
	// Optional custom endpoint (defaults to production)
	CDPEndpoint string

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
	Phone string `json:"phone,omitempty"`
	CioID string `json:"cio_id,omitempty"`
}

// IdentifyPayload represents the data for an identify call.
type IdentifyPayload struct {
	UserID string                 `json:"userId"`
	Traits map[string]interface{} `json:"traits"`
}

// TrackPayload represents the data for a track call.
type TrackPayload struct {
	UserID     string                 `json:"userId"`
	EventName  string                 `json:"event"`
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
	Attachments             map[string]string      `json:"attachments,omitempty"`
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

// DevicePayload represents data for registering a device.
type DevicePayload struct {
	UserID       string                 `json:"userId"`
	DeviceID     string                 `json:"device_id"`                // Unique device identifier (stable across token refreshes)
	Platform     string                 `json:"platform"`                 // "android", "ios", or "web"
	DeviceToken  string                 `json:"device_token"`             // FCM/APNS push token
	Name         string                 `json:"name,omitempty"`           // Device name
	OSVersion    string                 `json:"os_version,omitempty"`     // Operating system version
	Model        string                 `json:"model,omitempty"`          // Device model
	AppVersion   string                 `json:"app_version,omitempty"`    // Application version
	LastActiveAt string                 `json:"last_active_at,omitempty"` // Last active timestamp
	APNToken     string                 `json:"apn_token,omitempty"`      // Apple Push Notification token (iOS)
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}
