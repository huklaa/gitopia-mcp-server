package gitopia

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"sync"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

const DefaultGRPC = "gitopia-grpc.polkachu.com:11390"

type Client struct {
	mu        sync.RWMutex
	conn      *grpc.ClientConn
	Tm        tmservice.ServiceClient
	endpoints []string
	current   int
}

// GetConn returns the gRPC connection for external use
func (c *Client) GetConn() *grpc.ClientConn {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.conn
}

// transportCreds returns TLS credentials for port 443, insecure otherwise.
func transportCreds(endpoint string) grpc.DialOption {
	if strings.HasSuffix(endpoint, ":443") {
		return grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{}))
	}
	return grpc.WithTransportCredentials(insecure.NewCredentials())
}

// New creates a new Client connected to the first reachable endpoint.
func New(ctx context.Context, endpoints ...string) (*Client, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("at least one gRPC endpoint is required")
	}

	var lastErr error
	for i, ep := range endpoints {
		conn, err := grpc.DialContext(ctx, ep, transportCreds(ep))
		if err != nil {
			lastErr = err
			logging.Warnf("gRPC endpoint %s failed: %v", ep, err)
			continue
		}
		if i > 0 {
			logging.Infof("Connected to fallback gRPC endpoint: %s", ep)
		}
		return &Client{
			conn:      conn,
			Tm:        tmservice.NewServiceClient(conn),
			endpoints: endpoints,
			current:   i,
		}, nil
	}
	return nil, fmt.Errorf("all gRPC endpoints failed, last error: %w", lastErr)
}

// Reconnect cycles to the next available endpoint.
func (c *Client) Reconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.endpoints) <= 1 {
		return fmt.Errorf("no alternative endpoints available")
	}

	oldConn := c.conn
	var lastErr error

	for i := 1; i < len(c.endpoints); i++ {
		idx := (c.current + i) % len(c.endpoints)
		ep := c.endpoints[idx]
		conn, err := grpc.DialContext(ctx, ep, transportCreds(ep))
		if err != nil {
			lastErr = err
			logging.Warnf("Reconnect: gRPC endpoint %s failed: %v", ep, err)
			continue
		}
		c.conn = conn
		c.Tm = tmservice.NewServiceClient(conn)
		c.current = idx
		logging.Infof("Reconnected to gRPC endpoint: %s", ep)
		if oldConn != nil {
			_ = oldConn.Close()
		}
		return nil
	}
	return fmt.Errorf("reconnect failed, all endpoints exhausted, last error: %w", lastErr)
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Close()
}

// User represents a Gitopia user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Address  string `json:"address"`
}

// DAO represents a Gitopia DAO
type DAO struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	IsOwner bool   `json:"is_owner"`
}

// CreateUser signs & broadcasts a MsgCreateUser on-chain.
func (c *Client) CreateUser(ctx context.Context, w Wallet, username string) (*gitopiatypes.MsgCreateUserResponse, error) {
	msg := gitopiatypes.NewMsgCreateUser(
		w.Address(),
		username,
		"",  // name (optional)
		"",  // avatarUrl (optional)
		"",  // bio (optional)
	)

	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgCreateUserResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserByAddress queries user information by wallet address
func (c *Client) GetUserByAddress(ctx context.Context, address string) (*User, error) {
	q := gitopiatypes.NewQueryClient(c.conn)

	// Query user by address - this might need adjustment based on actual Gitopia API
	resp, err := q.User(ctx, &gitopiatypes.QueryGetUserRequest{
		Id: address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	// Find user by address
	if resp.User == nil {
		return nil, fmt.Errorf("user not found for address: %s", address)
	}
	return &User{
		ID:       strconv.FormatUint(resp.User.Id, 10),
		Username: resp.User.Username,
		Address:  resp.User.Creator,
	}, nil
}

// GetUserDAOs queries DAOs that a user is a member of
func (c *Client) GetUserDAOs(ctx context.Context, username string) ([]DAO, error) {
	q := gitopiatypes.NewQueryClient(c.conn)

	// Query all DAOs - this might need adjustment based on actual Gitopia API
	resp, err := q.UserDaoAll(ctx, &gitopiatypes.QueryAllUserDaoRequest{
		UserId: username,
		Pagination: &query.PageRequest{
			Limit: 1000,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query DAOs: %w", err)
	}

	var userDAOs []DAO
	for _, dao := range resp.Dao {
		// For now, we'll return all DAOs and let the caller filter
		// This logic needs to be adjusted based on actual DAO membership structure
		userDAOs = append(userDAOs, DAO{
			ID:      strconv.FormatUint(dao.Id, 10),
			Name:    dao.Name,
			Address: dao.Address,
			IsOwner: dao.Creator == username,
		})
	}

	return userDAOs, nil
}
