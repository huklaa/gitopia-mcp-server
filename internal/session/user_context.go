package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
)

// UserContext holds the current user session information
type UserContext struct {
	Username    string    `json:"username"`
	Address     string    `json:"address"`
	UserID      string    `json:"user_id"`
	DAOs        []DAO     `json:"daos,omitempty"`
	ActiveDAO   *DAO      `json:"active_dao,omitempty"`
	LastUpdated time.Time `json:"last_updated"`
}

// DAO represents a DAO that the user is a member of
type DAO struct {
	Name        string   `json:"name"`
	ID          string   `json:"id"`
	Address     string   `json:"address"`
	IsOwner     bool     `json:"is_owner"`
	Permissions []string `json:"permissions,omitempty"`
}

// ContextManager manages user context for MCP sessions
type ContextManager struct {
	mu      sync.RWMutex
	context *UserContext
	client  *gitopia.Client
}

// NewContextManager creates a new context manager
func NewContextManager(client *gitopia.Client) *ContextManager {
	return &ContextManager{
		client: client,
	}
}

// GetContext returns the current user context
func (cm *ContextManager) GetContext() *UserContext {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.context
}

// SetContext sets the user context
func (cm *ContextManager) SetContext(ctx *UserContext) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	ctx.LastUpdated = time.Now()
	cm.context = ctx
}

// InitializeFromWallet initializes user context from wallet address
func (cm *ContextManager) InitializeFromWallet(walletAddress string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Skip if already initialized for this address
	if cm.context != nil && cm.context.Address == walletAddress {
		return nil
	}

	logging.Infof("Initializing user context for wallet: %s", walletAddress)

	// Query user info from Gitopia
	user, err := cm.client.GetUserByAddress(context.Background(), walletAddress)
	if err != nil {
		return fmt.Errorf("failed to get user by address %s: %w", walletAddress, err)
	}

	// Query DAOs the user is a member of
	daos, err := cm.client.GetUserDAOs(context.Background(), user.Username)
	if err != nil {
		logging.Warnf("Failed to get DAOs for user %s: %v", user.Username, err)
		// Continue without DAOs - not critical
	}

	cm.context = &UserContext{
		Username:    user.Username,
		Address:     walletAddress,
		UserID:      user.ID,
		DAOs:        convertDAOs(daos),
		LastUpdated: time.Now(),
	}

	logging.Infof("User context initialized: %s (%s)", user.Username, walletAddress)
	return nil
}

// GetOwnerID returns the appropriate owner ID for operations
// If activeDAO is set, returns DAO ID, otherwise returns username
func (cm *ContextManager) GetOwnerID() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.context == nil {
		return ""
	}

	if cm.context.ActiveDAO != nil {
		return cm.context.ActiveDAO.ID
	}

	return cm.context.Username
}

// GetUsername returns the current username
func (cm *ContextManager) GetUsername() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.context == nil {
		return ""
	}

	return cm.context.Username
}

// SetActiveDAO sets the active DAO for operations
func (cm *ContextManager) SetActiveDAO(daoName string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.context == nil {
		return fmt.Errorf("no user context available")
	}

	if daoName == "" {
		cm.context.ActiveDAO = nil
		logging.Infof("Cleared active DAO, using user context: %s", cm.context.Username)
		return nil
	}

	// Find the DAO
	for i, dao := range cm.context.DAOs {
		if dao.Name == daoName {
			cm.context.ActiveDAO = &cm.context.DAOs[i]
			logging.Infof("Set active DAO: %s for user: %s", daoName, cm.context.Username)
			return nil
		}
	}

	return fmt.Errorf("DAO '%s' not found in user's DAOs", daoName)
}

// RefreshContext refreshes the user context from Gitopia
func (cm *ContextManager) RefreshContext() error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.context == nil {
		return fmt.Errorf("no context to refresh")
	}

	// Re-query user info
	user, err := cm.client.GetUserByAddress(context.Background(), cm.context.Address)
	if err != nil {
		return fmt.Errorf("failed to refresh user context: %w", err)
	}

	// Re-query DAOs
	daos, err := cm.client.GetUserDAOs(context.Background(), user.Username)
	if err != nil {
		logging.Warnf("Failed to refresh DAOs for user %s: %v", user.Username, err)
	} else {
		cm.context.DAOs = convertDAOs(daos)
	}

	cm.context.Username = user.Username
	cm.context.UserID = user.ID
	cm.context.LastUpdated = time.Now()

	logging.Infof("User context refreshed: %s", user.Username)
	return nil
}

// IsInitialized returns true if context is initialized
func (cm *ContextManager) IsInitialized() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.context != nil
}

// GetContextSummary returns a human-readable summary of the current context
func (cm *ContextManager) GetContextSummary() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.context == nil {
		return "No user context available"
	}

	summary := fmt.Sprintf("User: %s (%s)", cm.context.Username, cm.context.Address)

	if cm.context.ActiveDAO != nil {
		summary += fmt.Sprintf("\nActive DAO: %s", cm.context.ActiveDAO.Name)
	}

	if len(cm.context.DAOs) > 0 {
		summary += "\nAvailable DAOs: "
		for i, dao := range cm.context.DAOs {
			if i > 0 {
				summary += ", "
			}
			summary += dao.Name
		}
	}

	return summary
}

// convertDAOs converts Gitopia DAO objects to internal DAO structs
func convertDAOs(gitopiaDAOs []gitopia.DAO) []DAO {
	daos := make([]DAO, len(gitopiaDAOs))
	for i, dao := range gitopiaDAOs {
		daos[i] = DAO{
			Name:    dao.Name,
			ID:      dao.ID,
			Address: dao.Address,
			IsOwner: dao.IsOwner,
			// Add permissions if available in gitopia.DAO
		}
	}
	return daos
}

