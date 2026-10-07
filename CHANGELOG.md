# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `SendWhatsApp()` - Send WhatsApp messages using a saved WhatsApp transactional
- `Identifiers.CdpID` (`cdp_id`)
- Email attachments: `EmailPayload.Attach()` and `EmailPayload.AttachFile()` base64-encode files for you, and `SendEmail()` validates attachments against the gateway's limits (at most 5 files, 2 MB decoded in total) before sending

### Changed

- `Attachments` is now supported by the backend and no longer logs an "unsupported fields" warning
- Message sends (`/v1/send/*`) now fail over to a fallback gateway host only when the primary provably did not process the request (connection refused, DNS failure, HTTP 502/503). Timeouts, 4xx, 500 and 504 are returned without retrying, to avoid delivering the same message twice. Identify, track and device registration are unchanged.
- Message sends no longer follow HTTP redirects. A 3xx is returned as an error and is not retried on a fallback host, because the original host may already have accepted the message.
- Default gateway fallback hosts updated from `api.opencdp.com` / `api.opencdp.xyz` to `api.open-cdp.com` / `api.open-cdp.xyz` (primary remains `api.opencdp.io`)

### Deprecated

- `Identifiers.CioID` and `Identifiers.Phone`. The gateway never accepted them; the SDK now rejects them with a validation error instead of sending a request that fails with HTTP 400.


## [1.0.2] - 2026-01-29
### Added
- **MAJOR**: 14 new fields to `EmailPayload` model:
    - `BCC`, `CC` (email arrays)
    - `Preheader`, `AmpBody`, `PlaintextBody`
    - `Headers`, `Attachments`
    - `FakeBCC`, `DisableMessageRetention`, `SendToUnsubscribed`, `QueueDraft`, `DisableCSSPreprocessing`
    - `Language`
- Comprehensive email validation matching Node.js/PHP SDKs:
    - BCC/CC array validation (validates each email)
    - From/ReplyTo email validation
    - SendAt positive integer validation
    - Empty body field validation
    - Template vs raw email distinction
   - Raw emails require body, subject, and from fields
- "Exactly one" identifier validation for all messaging endpoints
    - Ensures only one of: id, email,or cio_id is provided
    - Validates email format in identifier.email field
- 8 comprehensive validation tests (26 total tests now passing)

### Changed
- **BREAKING**: Email validation is now comprehensive - invalid emails in BCC/CC/from/reply_to will be rejected
- **BREAKING**: Identifier validation now enforces "exactly one" (previously allowed multiple)

### Fixed
- **CRITICAL**: Fixed all API endpoints to match production API spec
  - Added `/v1` prefix to all endpoints (identify, track, registerDevice, ping, messaging)
  - Fixed default base URL from test server to production: `https://api.opencdp.io/gateway/data-gateway`
- Fixed properties validation to allow nil values (matches Node.js/PHP SDK behavior)
- Removed non-standard `ClearIdentity` method (not in reference SDK spec)

### Changed
- Updated all tests to expect correct API endpoints

## [1.0.0] - 2023-10-27
### Added
- Core SDK methods: `identify`, `track`, `sendEmail`, `sendPush`, `sendSms`, `ping`, and `registerDevice`.
- Customer.io dual-write support for seamless data migration.
- Configuration options: `timeout`, `failOnException`, `retryAttempts`, and `logger` customization.