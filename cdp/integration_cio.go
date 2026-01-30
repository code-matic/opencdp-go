package cdp

import (
	"github.com/customerio/go-customerio"
)

// CIOIntegration handles interactions with the Customer.io SDK.
type CIOIntegration struct {
	client  *customerio.CustomerIO
	enabled bool
}

// NewCIOIntegration creates a new Customer.io integration wrapper.
func NewCIOIntegration(config *CDPConfig) *CIOIntegration {
	if !config.SendToCustomerIO || config.CustomerIO == nil {
		return &CIOIntegration{enabled: false}
	}

	// Region handling could be added here if the underlying SDK supports it explicitly in the constructor,
	// but standard go-customerio usually takes siteID and apiKey.
	cio := customerio.NewCustomerIO(config.CustomerIO.SiteID, config.CustomerIO.APIKey)
	return &CIOIntegration{
		client:  cio,
		enabled: true,
	}
}

// Identify sends an identify call to Customer.io.
func (c *CIOIntegration) Identify(userID string, traits map[string]interface{}) error {
	if !c.enabled {
		return nil
	}
	return c.client.Identify(userID, traits)
}

// Track sends a track event to Customer.io.
func (c *CIOIntegration) Track(userID string, eventName string, properties map[string]interface{}) error {
	if !c.enabled {
		return nil
	}
	return c.client.Track(userID, eventName, properties)
}

// AddDevice registers a device for push notifications in Customer.io.
func (c *CIOIntegration) AddDevice(userID string, deviceID string, platform string, data map[string]interface{}) error {
	if !c.enabled {
		return nil
	}
	return c.client.AddDevice(userID, deviceID, platform, data)
}
