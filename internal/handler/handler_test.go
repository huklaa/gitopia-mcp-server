package handler

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	h := New(nil, nil, nil, nil)
	require.NotNil(t, h)
	assert.Nil(t, h.GClient)
	assert.Nil(t, h.GitClient)
	assert.Nil(t, h.AuthMgr)
	assert.Nil(t, h.WorkspaceMgr)
}

func TestRetryOnSequenceMismatch_Success(t *testing.T) {
	calls := 0
	result, err := retryOnSequenceMismatch(func() (int, error) {
		calls++
		return 42, nil
	})
	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 1, calls)
}

func TestRetryOnSequenceMismatch_NonSequenceError(t *testing.T) {
	calls := 0
	_, err := retryOnSequenceMismatch(func() (int, error) {
		calls++
		return 0, errors.New("some other error")
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some other error")
	assert.Equal(t, 1, calls) // Should not retry
}

func TestAuditLog_DoesNotPanic(t *testing.T) {
	// AuditLog should not panic on any input
	assert.NotPanics(t, func() {
		AuditLog("test_tool", "addr123", true, time.Millisecond*100, "")
	})
	assert.NotPanics(t, func() {
		AuditLog("test_tool", "addr123", false, time.Millisecond*50, "error detail")
	})
}
