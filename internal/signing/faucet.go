package signing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/signing/config"
)

const DefaultFaucetURL = "https://faucet.gitopia.com/"

// DefaultFeeGranter returns the configured fee granter address.
func DefaultFeeGranter() string {
	return config.FeeGranterAddr
}

// FeeGrantStatus reports the outcome of a fee grant claim attempt.
type FeeGrantStatus struct {
	Claimed bool   `json:"claimed"`
	Error   string `json:"error,omitempty"`
	Message string `json:"message,omitempty"`
}

// ClaimFeeGrant sends a POST request to the faucet to claim a fee grant for the given address.
// It is best-effort: on any failure (network, timeout, non-200) it returns a FeeGrantStatus
// with Claimed=false and an error description, but never a Go error.
func ClaimFeeGrant(ctx context.Context, address, faucetURL string) FeeGrantStatus {
	if faucetURL == "" {
		faucetURL = DefaultFaucetURL
	}

	body, err := json.Marshal(map[string]string{"address": address})
	if err != nil {
		return FeeGrantStatus{Error: fmt.Sprintf("failed to marshal request: %v", err)}
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, faucetURL, bytes.NewReader(body))
	if err != nil {
		return FeeGrantStatus{Error: fmt.Sprintf("failed to create request: %v", err)}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return FeeGrantStatus{Error: fmt.Sprintf("faucet request failed: %v", err)}
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	if resp.StatusCode != http.StatusOK {
		return FeeGrantStatus{
			Error:   fmt.Sprintf("faucet returned HTTP %d", resp.StatusCode),
			Message: string(respBody),
		}
	}

	return FeeGrantStatus{
		Claimed: true,
		Message: string(respBody),
	}
}
