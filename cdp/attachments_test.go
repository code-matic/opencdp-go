package cdp_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codematic/opencdp-go/cdp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pdfBase64 = base64.StdEncoding.EncodeToString([]byte("%PDF-1.4 test"))

func attachmentEmail(attachments map[string]string) cdp.EmailPayload {
	return cdp.EmailPayload{
		To:                     "test@example.com",
		Identifiers:            cdp.Identifiers{ID: "u1"},
		TransactionalMessageID: "INVOICE_EMAIL",
		Attachments:            attachments,
	}
}

func TestSendEmail_Attachments_SentWithoutWarning(t *testing.T) {
	var received map[string]interface{}
	server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn}))
	client := cdp.NewClient(mockConfig(server.URL, cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true, Logger: logger}))
	defer client.Close()

	err := client.SendEmail(context.Background(), attachmentEmail(map[string]string{"invoice.pdf": pdfBase64}))

	assert.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"invoice.pdf": pdfBase64}, received["attachments"])
	assert.NotContains(t, logs.String(), "unsupported")
}

func TestSendEmail_Attachments_Validation(t *testing.T) {
	sixFiles := map[string]string{}
	for i := 0; i < 6; i++ {
		sixFiles[fmt.Sprintf("file%d.txt", i)] = pdfBase64
	}
	tooBig := base64.StdEncoding.EncodeToString(make([]byte, 2*1024*1024+1))

	tests := []struct {
		name        string
		attachments map[string]string
		wantErr     string
	}{
		{"too many files", sixFiles, "attachments may contain at most 5 files"},
		{"over 2 MB decoded", map[string]string{"big.bin": tooBig}, "attachments decoded size exceeds 2097152 bytes (2 MB)"},
		{"path traversal", map[string]string{"../secret.pdf": pdfBase64}, "invalid attachment filename: ../secret.pdf"},
		{"forward slash", map[string]string{"dir/file.pdf": pdfBase64}, "invalid attachment filename: dir/file.pdf"},
		{"backslash", map[string]string{`dir\file.pdf`: pdfBase64}, `invalid attachment filename: dir\file.pdf`},
		{"empty filename", map[string]string{"": pdfBase64}, "invalid attachment filename: (empty)"},
		{"empty content", map[string]string{"a.pdf": ""}, `attachment "a.pdf" must be a non-empty base64 string`},
		{"not base64", map[string]string{"a.pdf": "!!!"}, `attachment "a.pdf" must be a valid base64 string`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			server := setupMockServer(t, func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})
			defer server.Close()

			client := cdp.NewClient(mockConfig(server.URL, cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true}))
			defer client.Close()

			err := client.SendEmail(context.Background(), attachmentEmail(tt.attachments))

			assert.ErrorContains(t, err, tt.wantErr)
			assert.False(t, called, "request must not be sent when attachments are invalid")
		})
	}
}

func TestSendEmail_Attachments_AcceptsUnpaddedAndURLSafeBase64(t *testing.T) {
	server := setupMockServer(t, defaultHandler(t, "/v1/send/email", "POST"))
	defer server.Close()

	client := cdp.NewClient(mockConfig(server.URL, cdp.CDPConfig{CDPAPIKey: "key", FailOnException: true}))
	defer client.Close()

	err := client.SendEmail(context.Background(), attachmentEmail(map[string]string{
		"unpadded.txt": "aGVsbG8",
		"urlsafe.bin":  base64.URLEncoding.EncodeToString([]byte{0xfb, 0xff}),
		"wrapped.txt":  "aGVs\nbG8=",
	}))

	assert.NoError(t, err)
}

func TestEmailPayload_Attach(t *testing.T) {
	payload := attachmentEmail(map[string]string{"a.txt": "YQ=="})

	payload.Attach("invoice.pdf", []byte("%PDF-1.4 test"))

	assert.Equal(t, map[string]string{"a.txt": "YQ==", "invoice.pdf": pdfBase64}, payload.Attachments)
}

func TestEmailPayload_AttachFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "invoice.pdf")
	require.NoError(t, os.WriteFile(path, []byte("%PDF-1.4 test"), 0o600))
	payload := attachmentEmail(nil)

	require.NoError(t, payload.AttachFile(path))

	assert.Equal(t, map[string]string{"invoice.pdf": pdfBase64}, payload.Attachments)
}

func TestEmailPayload_AttachFile_MissingFile(t *testing.T) {
	payload := attachmentEmail(nil)

	err := payload.AttachFile(filepath.Join(t.TempDir(), "missing.pdf"))

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "missing.pdf"))
	assert.Nil(t, payload.Attachments)
}
