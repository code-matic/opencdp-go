# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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