package session

import (
	"sync"
	"testing"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextManager(t *testing.T) {
	cm := NewContextManager(nil)
	require.NotNil(t, cm)
	assert.False(t, cm.IsInitialized())
}

func TestSetAndGetContext(t *testing.T) {
	cm := NewContextManager(nil)

	before := time.Now()
	ctx := &UserContext{
		Username: "alice",
		Address:  "gitopia1abc",
		UserID:   "42",
	}
	cm.SetContext(ctx)

	got := cm.GetContext()
	require.NotNil(t, got)
	assert.Equal(t, "alice", got.Username)
	assert.Equal(t, "gitopia1abc", got.Address)
	assert.Equal(t, "42", got.UserID)
	assert.False(t, got.LastUpdated.Before(before))
}

func TestGetOwnerID_NilContext(t *testing.T) {
	cm := NewContextManager(nil)
	assert.Equal(t, "", cm.GetOwnerID())
}

func TestGetOwnerID_NoActiveDAO(t *testing.T) {
	cm := NewContextManager(nil)
	cm.SetContext(&UserContext{Username: "bob"})
	assert.Equal(t, "bob", cm.GetOwnerID())
}

func TestGetOwnerID_WithActiveDAO(t *testing.T) {
	cm := NewContextManager(nil)
	dao := DAO{Name: "myorg", ID: "dao-123"}
	cm.SetContext(&UserContext{
		Username:  "bob",
		DAOs:      []DAO{dao},
		ActiveDAO: &dao,
	})
	assert.Equal(t, "myorg", cm.GetOwnerID())
}

func TestGetUsername_NilContext(t *testing.T) {
	cm := NewContextManager(nil)
	assert.Equal(t, "", cm.GetUsername())
}

func TestGetUsername_WithContext(t *testing.T) {
	cm := NewContextManager(nil)
	cm.SetContext(&UserContext{Username: "carol"})
	assert.Equal(t, "carol", cm.GetUsername())
}

func TestSetActiveDAO_NilContext(t *testing.T) {
	cm := NewContextManager(nil)
	err := cm.SetActiveDAO("myorg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no user context")
}

func TestSetActiveDAO_ClearDAO(t *testing.T) {
	cm := NewContextManager(nil)
	dao := DAO{Name: "myorg", ID: "dao-1"}
	cm.SetContext(&UserContext{
		Username:  "alice",
		DAOs:      []DAO{dao},
		ActiveDAO: &dao,
	})
	require.NotNil(t, cm.GetContext().ActiveDAO)

	err := cm.SetActiveDAO("")
	assert.NoError(t, err)
	assert.Nil(t, cm.GetContext().ActiveDAO)
}

func TestSetActiveDAO_Found(t *testing.T) {
	cm := NewContextManager(nil)
	cm.SetContext(&UserContext{
		Username: "alice",
		DAOs: []DAO{
			{Name: "org-a", ID: "1"},
			{Name: "org-b", ID: "2"},
		},
	})

	err := cm.SetActiveDAO("org-b")
	assert.NoError(t, err)
	require.NotNil(t, cm.GetContext().ActiveDAO)
	assert.Equal(t, "2", cm.GetContext().ActiveDAO.ID)
}

func TestSetActiveDAO_NotFound(t *testing.T) {
	cm := NewContextManager(nil)
	cm.SetContext(&UserContext{
		Username: "alice",
		DAOs:     []DAO{{Name: "org-a", ID: "1"}},
	})

	err := cm.SetActiveDAO("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIsInitialized(t *testing.T) {
	cm := NewContextManager(nil)
	assert.False(t, cm.IsInitialized())

	cm.SetContext(&UserContext{Username: "alice"})
	assert.True(t, cm.IsInitialized())
}

func TestGetContextSummary_NilContext(t *testing.T) {
	cm := NewContextManager(nil)
	assert.Equal(t, "No user context available", cm.GetContextSummary())
}

func TestGetContextSummary_WithDAOs(t *testing.T) {
	cm := NewContextManager(nil)
	cm.SetContext(&UserContext{
		Username: "alice",
		Address:  "gitopia1xyz",
		DAOs: []DAO{
			{Name: "dao-one"},
			{Name: "dao-two"},
		},
	})

	summary := cm.GetContextSummary()
	assert.Contains(t, summary, "alice")
	assert.Contains(t, summary, "gitopia1xyz")
	assert.Contains(t, summary, "dao-one")
	assert.Contains(t, summary, "dao-two")
}

func TestGetContextSummary_WithActiveDAO(t *testing.T) {
	cm := NewContextManager(nil)
	dao := DAO{Name: "active-dao"}
	cm.SetContext(&UserContext{
		Username:  "alice",
		Address:   "gitopia1xyz",
		ActiveDAO: &dao,
	})

	summary := cm.GetContextSummary()
	assert.Contains(t, summary, "Active DAO: active-dao")
}

func TestConvertDAOs(t *testing.T) {
	input := []gitopia.DAO{
		{Name: "dao1", ID: "id1", Address: "addr1", IsOwner: true},
		{Name: "dao2", ID: "id2", Address: "addr2", IsOwner: false},
	}

	result := convertDAOs(input)
	require.Len(t, result, 2)

	assert.Equal(t, "dao1", result[0].Name)
	assert.Equal(t, "id1", result[0].ID)
	assert.Equal(t, "addr1", result[0].Address)
	assert.True(t, result[0].IsOwner)

	assert.Equal(t, "dao2", result[1].Name)
	assert.Equal(t, "id2", result[1].ID)
	assert.False(t, result[1].IsOwner)
}

func TestConvertDAOs_Empty(t *testing.T) {
	result := convertDAOs(nil)
	assert.Empty(t, result)
}

func TestConcurrentAccess(t *testing.T) {
	cm := NewContextManager(nil)
	cm.SetContext(&UserContext{
		Username: "concurrent-user",
		DAOs:     []DAO{{Name: "org", ID: "1"}},
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = cm.GetContext()
		}()
		go func() {
			defer wg.Done()
			cm.SetContext(&UserContext{Username: "concurrent-user", DAOs: []DAO{{Name: "org", ID: "1"}}})
		}()
		go func() {
			defer wg.Done()
			_ = cm.GetOwnerID()
		}()
	}
	wg.Wait()
}
