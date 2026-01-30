package cdp

import (
	"context"
	"strings"
)

// SendEmail sends a transactional or raw email.
func (c *Client) SendEmail(ctx context.Context, payload EmailPayload) error {
	// Comprehensive validation
	if err := validateSendEmailRequest(payload); err != nil {
		return c.handleError(ctx, err, "Validation failed for SendEmail")
	}

	// Check for unsupported fields and warn (mirrors Node.js SDK)
	var unsupported []string
	if payload.SendAt != 0 {
		unsupported = append(unsupported, "send_at")
	}
	if payload.Tracked {
		unsupported = append(unsupported, "tracked")
	}
	if payload.DisableMessageRetention {
		unsupported = append(unsupported, "disable_message_retention")
	}
	if payload.SendToUnsubscribed {
		unsupported = append(unsupported, "send_to_unsubscribed")
	}
	if payload.QueueDraft {
		unsupported = append(unsupported, "queue_draft")
	}
	if len(payload.Headers) > 0 {
		unsupported = append(unsupported, "headers")
	}
	if payload.DisableCSSPreprocessing {
		unsupported = append(unsupported, "disable_css_preprocessing")
	}
	if payload.FakeBCC {
		unsupported = append(unsupported, "fake_bcc")
	}
	if payload.ReplyTo != "" {
		unsupported = append(unsupported, "reply_to")
	}
	if payload.Preheader != "" {
		unsupported = append(unsupported, "preheader")
	}
	if len(payload.Attachments) > 0 {
		unsupported = append(unsupported, "attachments")
	}
	if len(unsupported) > 0 {
		c.logger.Warn("SendEmail contains unsupported fields which will be ignored: " + strings.Join(unsupported, ", ") + ". These are included for future compatibility.")
	}

	// Warn about dual-write behavior
	if c.config.SendToCustomerIO {
		c.logger.Warn("Transactional email will NOT be sent to Customer.io to avoid duplicates. Set SendToCustomerIO to false to suppress this warning.")
	}

	c.logger.Debug("Sending email", "to", payload.To)

	if err := c.post(ctx, "/v1/send/email", payload); err != nil {
		return c.handleError(ctx, NewCDPEmailError("Failed to send email", err), "SendEmail failed")
	}

	return nil
}

// SendPush sends a push notification.
func (c *Client) SendPush(ctx context.Context, payload PushPayload) error {
	// Validate identifiers - must have exactly one
	if err := validateSendPushRequest(payload); err != nil {
		return c.handleError(ctx, err, "Validation failed for SendPush")
	}

	// Warn about dual-write behavior
	if c.config.SendToCustomerIO {
		c.logger.Warn("Transactional push will NOT be sent to Customer.io to avoid duplicates. Set SendToCustomerIO to false to suppress this warning.")
	}

	c.logger.Debug("Sending push notification", "transactional_message_id", payload.TransactionalMessageID)

	if err := c.post(ctx, "/v1/send/push", payload); err != nil {
		return c.handleError(ctx, NewCDPPushError("Failed to send push", err), "SendPush failed")
	}

	return nil
}

// SendSms sends an SMS message.
func (c *Client) SendSms(ctx context.Context, payload SmsPayload) error {
	// Comprehensive validation
	if err := validateSendSmsRequest(payload); err != nil {
		return c.handleError(ctx, err, "Validation failed for SendSms")
	}

	c.logger.Debug("Sending SMS")

	if err := c.post(ctx, "/v1/send/sms", payload); err != nil {
		return c.handleError(ctx, NewCDPSmsError("Failed to send SMS", err), "SendSms failed")
	}

	return nil
}
