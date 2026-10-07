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

### Email Attachments

Attach up to 5 files (2 MB decoded in total). `Attach` and `AttachFile` base64-encode the content for you:

```go
payload := cdp.EmailPayload{
    To: "alice@example.com",
    Identifiers: cdp.Identifiers{ID: "user-123"},
    TransactionalMessageID: "INVOICE_EMAIL",
}
if err := payload.AttachFile("./invoice.pdf"); err != nil { // named "invoice.pdf"
    return err
}
payload.Attach("notes.txt", []byte("Plain text content"))

client.SendEmail(ctx, payload)
```

Content that is already base64 can be set on `payload.Attachments` (filename to base64) directly. `SendEmail` validates attachments before sending: at most 5 files, at most 2 MB decoded in total, filenames without `/`, `\` or `..`, and non-empty base64 content. The content type is inferred from the file extension.

### Send WhatsApp

WhatsApp sends use a saved WhatsApp transactional, so `TransactionalMessageID` is required.

```go
err := client.SendWhatsApp(ctx, cdp.WhatsAppPayload{
    Identifiers:            cdp.Identifiers{ID: "user-123"},
    TransactionalMessageID: "ORDER_WHATSAPP",
    To:                     "+14155551234", // Optional: overrides the profile phone number
    MessageData:            map[string]interface{}{"order_number": "12345"}, // {{trigger.order_number}} in the template
})
```

- **A nil error means the message was queued, not delivered.** Delivery runs asynchronously, so a missing WhatsApp provider, no phone number, or a template rejected by Meta does not fail this call.
- `TemplateVariables` (`&cdp.WhatsAppTemplateVars{Header: ..., Body: ..., Button: ...}`) sets the template slots from code. Keys must be slot numbers (`"1"`, `"2"`, ...), and values may use Liquid such as `{{customer.first_name}}`. Setting it **replaces all variables saved on the transactional**, so include every section the template needs.
- Identify the user with exactly one of `ID`, `Email`, or `CdpID`.
- Sends are not retried on another gateway host after a timeout or an HTTP error, because the message may already have been queued. The exceptions are connection failures and the Cloudflare errors 521, 522, 523, 525 and 526, which mean the gateway never received the request.

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
