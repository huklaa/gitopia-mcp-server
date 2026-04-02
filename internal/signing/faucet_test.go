package signing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaimFeeGrant_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	status := ClaimFeeGrant(context.Background(), "gitopia1abc123", srv.URL)
	assert.True(t, status.Claimed)
	assert.Empty(t, status.Error)
	assert.Contains(t, status.Message, "ok")
}

func TestClaimFeeGrant_FaucetDown(t *testing.T) {
	// Use an unreachable URL
	status := ClaimFeeGrant(context.Background(), "gitopia1abc123", "http://127.0.0.1:1")
	assert.False(t, status.Claimed)
	assert.NotEmpty(t, status.Error)
	assert.Contains(t, status.Error, "faucet request failed")
}

func TestClaimFeeGrant_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`rate limited`))
	}))
	defer srv.Close()

	status := ClaimFeeGrant(context.Background(), "gitopia1abc123", srv.URL)
	assert.False(t, status.Claimed)
	assert.Contains(t, status.Error, "429")
	assert.Contains(t, status.Message, "rate limited")
}
