package cdp_test

import (
	"testing"

	"github.com/codematic/opencdp-go/cdp"
	"github.com/stretchr/testify/assert"
)

func TestResolveAllBaseURLs_DefaultPrimaryAndFallbacks(t *testing.T) {
	urls := cdp.ResolveAllBaseURLs("", nil)
	assert.Equal(t, "https://api.opencdp.io/gateway/data-gateway", urls[0])
	assert.Equal(t, []string{
		"https://api.opencdp.io/gateway/data-gateway",
		"https://api.open-cdp.com/gateway/data-gateway",
		"https://api.open-cdp.xyz/gateway/data-gateway",
	}, urls)
}

func TestResolveAllBaseURLs_EmptyFallbackSliceUsesPrimaryOnly(t *testing.T) {
	primary := "https://mock.test/gateway"
	urls := cdp.ResolveAllBaseURLs(primary, []string{})
	assert.Equal(t, []string{primary}, urls)
}

func TestResolveAllBaseURLs_DeduplicatesPrimaryAndFallbacks(t *testing.T) {
	primary := "https://api.opencdp.io/gateway/data-gateway"
	urls := cdp.ResolveAllBaseURLs(primary, []string{
		"https://api.open-cdp.com/gateway/data-gateway",
		primary,
	})
	assert.Equal(t, []string{
		"https://api.opencdp.io/gateway/data-gateway",
		"https://api.open-cdp.com/gateway/data-gateway",
	}, urls)
}
