package cdp

import "fmt"

// CDPError is the base error type for the SDK.
type CDPError struct {
	Message string
	Code    string
	Cause   error
}

func (e *CDPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewCDPError creates a new CDPError.
func NewCDPError(code, message string, cause error) *CDPError {
	return &CDPError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// CDPEmailError represents an error sending email.
type CDPEmailError struct {
	*CDPError
}

// NewCDPEmailError creates a new CDPEmailError.
func NewCDPEmailError(message string, cause error) *CDPEmailError {
	return &CDPEmailError{NewCDPError("EMAIL_ERROR", message, cause)}
}

// CDPPushError represents an error sending push notifications.
type CDPPushError struct {
	*CDPError
}

// NewCDPPushError creates a new CDPPushError.
func NewCDPPushError(message string, cause error) *CDPPushError {
	return &CDPPushError{NewCDPError("PUSH_ERROR", message, cause)}
}

// CDPSmsError represents an error sending SMS.
type CDPSmsError struct {
	*CDPError
}

// NewCDPSmsError creates a new CDPSmsError.
func NewCDPSmsError(message string, cause error) *CDPSmsError {
	return &CDPSmsError{NewCDPError("SMS_ERROR", message, cause)}
}

// CDPValidationError represents an input validation error.
type CDPValidationError struct {
	*CDPError
}

// NewCDPValidationError creates a new CDPValidationError.
func NewCDPValidationError(message string) *CDPValidationError {
	return &CDPValidationError{NewCDPError("VALIDATION_ERROR", message, nil)}
}
