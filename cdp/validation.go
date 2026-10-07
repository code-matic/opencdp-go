package cdp

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	emailRegex   = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	phoneRegex   = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	slotKeyRegex = regexp.MustCompile(`^[1-9]\d*$`)
)

// validateIdentifier checks if the identifier is non-empty.
func validateIdentifier(id string) error {
	if id == "" {
		return NewCDPValidationError("identifier cannot be empty")
	}
	return nil
}

// validateEmail checks if the email format is valid.
func validateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return NewCDPValidationError(fmt.Sprintf("invalid email format: %s", email))
	}
	return nil
}

// validatePhoneNumber checks if the phone number is in E.164 format.
func validatePhoneNumber(phone string) error {
	if !phoneRegex.MatchString(phone) {
		return NewCDPValidationError(fmt.Sprintf("invalid phone number format (E.164 required): %s", phone))
	}
	return nil
}

// validateEventName checks if the event name is non-empty.
func validateEventName(name string) error {
	if name == "" {
		return NewCDPValidationError("event name cannot be empty")
	}
	return nil
}

// validateProperties checks if the properties map is valid.
// Allows nil (will default to empty map in JSON serialization).
func validateProperties(props map[string]interface{}) error {
	// nil is allowed - it will be serialized as empty object
	return nil
}

// validateIdentifiers validates that exactly one identifier is provided.
func validateIdentifiers(identifiers Identifiers) error {
	count := 0
	hasID := identifiers.ID != ""
	hasEmail := identifiers.Email != ""
	hasCdpID := identifiers.CdpID != ""

	if identifiers.CioID != "" || identifiers.Phone != "" {
		return NewCDPValidationError("identifiers.cio_id and identifiers.phone are not supported; use id, email, or cdp_id")
	}

	if hasID {
		count++
	}
	if hasEmail {
		count++
		if err := validateEmail(identifiers.Email); err != nil {
			return err
		}
	}
	if hasCdpID {
		count++
	}

	if count == 0 {
		return NewCDPValidationError("identifiers must contain exactly one of: id, email, or cdp_id")
	}
	if count > 1 {
		return NewCDPValidationError("identifiers must contain exactly one of: id, email, or cdp_id (found multiple)")
	}

	return nil
}

// validateSendEmailRequest validates the email request comprehensively.
func validateSendEmailRequest(payload EmailPayload) error {
	// Validate required fields
	if payload.To == "" {
		return NewCDPValidationError("to is required")
	}
	if err := validateEmail(payload.To); err != nil {
		return err
	}

	// Validate identifiers - must contain exactly one
	if err := validateIdentifiers(payload.Identifiers); err != nil {
		return err
	}

	// Validate from if provided
	if payload.From != "" {
		if err := validateEmail(payload.From); err != nil {
			return err
		}
	}

	// Validate BCC emails
	if len(payload.BCC) > 0 {
		for _, email := range payload.BCC {
			if err := validateEmail(email); err != nil {
				return fmt.Errorf("invalid bcc email: %w", err)
			}
		}
	}

	// Validate CC emails
	if len(payload.CC) > 0 {
		for _, email := range payload.CC {
			if err := validateEmail(email); err != nil {
				return fmt.Errorf("invalid cc email: %w", err)
			}
		}
	}

	// Validate reply_to
	if payload.ReplyTo != "" {
		if err := validateEmail(payload.ReplyTo); err != nil {
			return err
		}
	}

	// Validate send_at if provided
	if payload.SendAt != 0 {
		if payload.SendAt < 0 {
			return NewCDPValidationError("send_at must be a positive integer")
		}
	}

	// Validate body fields - cannot be empty strings if provided
	if payload.Body != "" && strings.TrimSpace(payload.Body) == "" {
		return NewCDPValidationError("body cannot be empty if provided")
	}
	if payload.AmpBody != "" && strings.TrimSpace(payload.AmpBody) == "" {
		return NewCDPValidationError("amp_body cannot be empty if provided")
	}
	if payload.PlaintextBody != "" && strings.TrimSpace(payload.PlaintextBody) == "" {
		return NewCDPValidationError("plaintext_body cannot be empty if provided")
	}

	// Type guard to check if it's a template-based request
	isTemplateRequest := payload.TransactionalMessageID != ""

	if !isTemplateRequest {
		// Raw email - body, subject, and from are required
		errors := []string{}

		if payload.Body == "" {
			errors = append(errors, "body is required when not using a template")
		}
		if payload.Subject == "" {
			errors = append(errors, "subject is required when not using a template")
		}
		if payload.From == "" {
			errors = append(errors, "from is required when not using a template")
		}

		if len(errors) > 0 {
			return NewCDPValidationError(fmt.Sprintf("When not using a template: %s", strings.Join(errors, ", ")))
		}
	}

	return nil
}

// validateSendPushRequest validates the push notification request.
func validateSendPushRequest(payload PushPayload) error {
	// Validate identifiers - must contain exactly one
	if err := validateIdentifiers(payload.Identifiers); err != nil {
		return err
	}

	// TransactionalMessageID is required (mirrors Node.js SDK)
	if payload.TransactionalMessageID == "" {
		return NewCDPValidationError("transactional_message_id is required")
	}

	// Validate body field - cannot be empty string if provided
	if payload.Body != "" && strings.TrimSpace(payload.Body) == "" {
		return NewCDPValidationError("body cannot be empty if provided")
	}

	return nil
}

// validateSendSmsRequest validates the SMS request.
func validateSendSmsRequest(payload SmsPayload) error {
	// Validate identifiers - must contain exactly one
	if err := validateIdentifiers(payload.Identifiers); err != nil {
		return err
	}

	// Validate conditional requirement: body is required if no transactional_message_id (mirrors Node.js SDK)
	hasTemplateID := payload.TransactionalMessageID != ""
	if !hasTemplateID && payload.Body == "" {
		return NewCDPValidationError("body is required when not using a template")
	}

	// Validate phone number format if to is provided
	if payload.To != "" {
		if err := validatePhoneNumber(payload.To); err != nil {
			return err
		}
	}

	// Validate from phone number format if provided
	if payload.From != "" {
		if err := validatePhoneNumber(payload.From); err != nil {
			return err
		}
	}

	// Validate body field - cannot be empty string if provided
	if payload.Body != "" && strings.TrimSpace(payload.Body) == "" {
		return NewCDPValidationError("body cannot be empty if provided")
	}

	return nil
}

// validateSendWhatsAppRequest validates the WhatsApp request.
func validateSendWhatsAppRequest(payload WhatsAppPayload) error {
	if err := validateIdentifiers(payload.Identifiers); err != nil {
		return err
	}

	if payload.TransactionalMessageID == "" {
		return NewCDPValidationError("transactional_message_id is required")
	}

	if payload.To != "" {
		if err := validatePhoneNumber(payload.To); err != nil {
			return err
		}
	}

	if vars := payload.TemplateVariables; vars != nil {
		sections := []struct {
			name  string
			slots map[string]interface{}
		}{{"header", vars.Header}, {"body", vars.Body}, {"button", vars.Button}}
		for _, section := range sections {
			// The gateway sends parameters by position and drops non-numeric button keys.
			for slot := range section.slots {
				if !slotKeyRegex.MatchString(slot) {
					return NewCDPValidationError(fmt.Sprintf(
						"template_variables.%s keys must be positional slot numbers (\"1\", \"2\", ...), got %q",
						section.name, slot))
				}
			}
		}
	}

	return nil
}
