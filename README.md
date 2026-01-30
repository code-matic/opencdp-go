# CDP Go SDK

A Go client library for Codematic's Customer Data Platform (CDP) with optional Customer.io integration.

## Installation

```bash
go get github.com/codematic/opencdp-go
```

## Usage

### Initialization

```go
import "github.com/codematic/opencdp-go/cdp"

config := cdp.CDPConfig{
    CDPAPIKey:       "your-api-key",
    Debug:           true,
    FailOnException: true,
    Timeout:         10000,
}

client := cdp.NewClient(config)
defer client.Close()
```

### Identify a User

```go
ctx := context.Background()
traits := map[string]interface{}{
    "name": "Alice",
    "email": "alice@example.com",
}
err := client.Identify(ctx, "user-123", traits)
```

### Track an Event

```go
props := map[string]interface{}{
    "sku": "12345",
    "value": 100,
}
err := client.Track(ctx, "user-123", "purchase", props)
```

### Send Messaging

```go
// Email
client.SendEmail(ctx, cdp.EmailPayload{
    To: "alice@example.com",
    Identifiers: cdp.Identifiers{ID: "user-123"},
    TransactionalMessageID: "WELCOME",
})

// Push
client.SendPush(ctx, cdp.PushPayload{
    Identifiers: cdp.Identifiers{ID: "user-123"},
    TransactionalMessageID: "PUSH_PROMO",
    Title: "Sale!",
})
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `CDPAPIKey` | string | Required | Your CDP API Key |
| `Timeout` | int | 10000 | Request timeout in ms |
| `FailOnException` | bool | false | If true, returns errors; else logs and returns nil |
| `MaxConcurrentRequests` | int | 10 | Max concurrent HTTP requests (max 30) |
| `Debug` | bool | false | Enable debug logging |
| `SendToCustomerIO` | bool | false | Enable dual-write to Customer.io |

## Dual-Write to Customer.io

To enable dual-write, configure the `CustomerIO` struct in the config:

```go
config := cdp.CDPConfig{
    CDPAPIKey: "...",
    SendToCustomerIO: true,
    CustomerIO: &cdp.CustomerIOConfig{
        SiteID: "cio-site-id",
        APIKey: "cio-api-key",
    },
}
```
