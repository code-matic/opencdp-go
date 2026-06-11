package cdp

import "strings"

const (
	defaultPrimaryBaseURL = "https://api.opencdp.com/gateway/data-gateway"
	backupBaseURLXyz      = "https://api.opencdp.xyz/gateway/data-gateway"
	backupBaseURLIo       = "https://api.opencdp.io/gateway/data-gateway"
)

var defaultFallbackBaseURLs = []string{backupBaseURLXyz, backupBaseURLIo}

func normalizeBaseURL(url string) string {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return trimmed
	}
	return strings.TrimSuffix(trimmed, "/")
}

// ResolveAllBaseURLs returns ordered gateway roots: primary first, then fallbacks.
func ResolveAllBaseURLs(primaryOverride string, fallbackOverrides []string) []string {
	primary := normalizeBaseURL(primaryOverride)
	if primary == "" {
		primary = defaultPrimaryBaseURL
	}
	fallbacks := fallbackOverrides
	if fallbacks == nil {
		fallbacks = defaultFallbackBaseURLs
	}
	seen := make(map[string]struct{})
	ordered := make([]string, 0, 1+len(fallbacks))
	add := func(url string) {
		normalized := normalizeBaseURL(url)
		if normalized == "" {
			return
		}
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		ordered = append(ordered, normalized)
	}
	add(primary)
	for _, fb := range fallbacks {
		add(fb)
	}
	return ordered
}
