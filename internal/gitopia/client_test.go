package gitopia

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_NoEndpoints(t *testing.T) {
	_, err := New(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one gRPC endpoint is required")
}

func TestNew_SingleEndpoint(t *testing.T) {
	// gRPC DialContext is lazy by default (no WithBlock), so even an
	// unreachable endpoint will return a client without error.
	client, err := New(context.Background(), "localhost:0")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.NotNil(t, client.GetConn())
	assert.Equal(t, 0, client.current)
	assert.Len(t, client.endpoints, 1)
	defer func() { _ = client.Close() }()
}

func TestNew_MultipleEndpoints(t *testing.T) {
	client, err := New(context.Background(), "localhost:0", "localhost:1")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, 0, client.current)
	assert.Len(t, client.endpoints, 2)
	defer func() { _ = client.Close() }()
}

func TestReconnect_SingleEndpoint(t *testing.T) {
	client, err := New(context.Background(), "localhost:0")
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	err = client.Reconnect(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no alternative endpoints")
}

func TestReconnect_CyclesToNext(t *testing.T) {
	client, err := New(context.Background(), "localhost:0", "localhost:1")
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	err = client.Reconnect(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, client.current)
	// Verify connection object exists after reconnect.
	// Note: we don't compare old vs new conn pointers because gRPC
	// background goroutines mutate internal conn state, causing data races.
	assert.NotNil(t, client.GetConn())
}

func TestReconnect_WrapsAround(t *testing.T) {
	client, err := New(context.Background(), "localhost:0", "localhost:1", "localhost:2")
	require.NoError(t, err)
	defer func() { _ = client.Close() }()

	// Reconnect from 0 → should go to 1
	require.NoError(t, client.Reconnect(context.Background()))
	assert.Equal(t, 1, client.current)

	// Reconnect from 1 → should go to 2
	require.NoError(t, client.Reconnect(context.Background()))
	assert.Equal(t, 2, client.current)

	// Reconnect from 2 → should wrap to 0
	require.NoError(t, client.Reconnect(context.Background()))
	assert.Equal(t, 0, client.current)
}

func TestClose(t *testing.T) {
	client, err := New(context.Background(), "localhost:0")
	require.NoError(t, err)
	assert.NoError(t, client.Close())
}
