package gitopia

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/gitopia/gitopia-mcp-server/internal/signing"
	"github.com/gitopia/gitopia-mcp-server/internal/signing/config"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
)

// ListUserRepos returns up to `limit` repos owned by an address / username / DAO name.
func (c *Client) ListUserRepos(
	ctx context.Context,
	owner string,
	limit uint64,
) ([]*gitopiatypes.Repository, error) {

	q := gitopiatypes.NewQueryClient(c.conn)

	resp, err := q.AnyRepositoryAll(ctx, &gitopiatypes.QueryAllAnyRepositoryRequest{
		Id: owner,
		Pagination: &query.PageRequest{
			Limit: limit,
		},
	})
	if err != nil {
		return nil, err
	}
	return resp.Repository, nil
}

// CreateRepository signs & broadcasts a MsgCreateRepository and then
// returns the repository URL instead of querying the chain immediately.
func (c *Client) CreateRepository(
	ctx context.Context,
	w Wallet, // wallet able to SignAndBroadcast
	ownerId string, // bech32 address of USER or DAO
	name, description string,
) (*gitopiatypes.MsgCreateRepositoryResponse, error) {

	msg := gitopiatypes.NewMsgCreateRepository(
		w.Address(), // creator
		name,
		ownerId,
		description,
	)

	// Broadcast the transaction
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgCreateRepositoryResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// GetRepository returns a single repository by owner and name.
func (c *Client) GetRepository(
	ctx context.Context,
	owner, name string,
) (*gitopiatypes.Repository, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.AnyRepository(ctx, &gitopiatypes.QueryGetAnyRepositoryRequest{
		Id:             owner,
		RepositoryName: name,
	})
	if err != nil {
		return nil, err
	}
	return resp.Repository, nil
}

func (c *Client) ListBranches(
	ctx context.Context,
	ownerId, repoName string,
	limit uint64,
) ([]*gitopiatypes.Branch, error) {

	q := gitopiatypes.NewQueryClient(c.conn)

	resp, err := q.RepositoryBranchAll(ctx, &gitopiatypes.QueryAllRepositoryBranchRequest{
		Id:             ownerId,
		RepositoryName: repoName,
		Pagination: &query.PageRequest{
			Limit: limit,
		},
	})
	if err != nil {
		return nil, err
	}
	branches := make([]*gitopiatypes.Branch, len(resp.Branch))
	for i := range resp.Branch {
		branches[i] = &resp.Branch[i]
	}
	return branches, nil
}

// GetFile fetches the raw content of a file from a repository.
func (c *Client) GetFile(
	ctx context.Context,
	ownerId, repoName, branch, path string,
) ([]byte, error) {
	// The PRD suggests file content is available via an IPFS gateway. We construct
	// the URL based on the pattern observed for public git hosts.
	// The host is loaded from the git config or defaults.
	_ = config.LoadGitConfig()
	url := fmt.Sprintf("%s/raw/%s/%s/%s/%s",
		config.GitServerHost, ownerId, repoName, branch, path)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file from %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from %s: %s", url, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	return body, nil
}

// ListIssues returns all issues for a repository.
func (c *Client) ListIssues(
	ctx context.Context,
	ownerId, repoName string,
	limit uint64,
) ([]*gitopiatypes.Issue, error) {

	q := gitopiatypes.NewQueryClient(c.conn)

	resp, err := q.RepositoryIssueAll(ctx, &gitopiatypes.QueryAllRepositoryIssueRequest{
		Id:             ownerId,
		RepositoryName: repoName,
		Pagination: &query.PageRequest{
			Limit: limit,
		},
	})
	if err != nil {
		return nil, err
	}
	return resp.Issue, nil
}

// GetIssue returns a single issue by owner, repo name, and issue IID.
func (c *Client) GetIssue(
	ctx context.Context,
	ownerId, repoName string,
	issueIid uint64,
) (*gitopiatypes.Issue, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.RepositoryIssue(ctx, &gitopiatypes.QueryGetRepositoryIssueRequest{
		Id:             ownerId,
		RepositoryName: repoName,
		IssueIid:       issueIid,
	})
	if err != nil {
		return nil, err
	}
	return resp.Issue, nil
}

// CreateIssue signs & broadcasts a MsgCreateIssue.
func (c *Client) CreateIssue(
	ctx context.Context,
	w Wallet, // wallet able to SignAndBroadcast
	ownerId, repoName, title, description string,
	assignees []string,
	labelIDs []uint64,
) (string, error) {

	msg := gitopiatypes.NewMsgCreateIssue(
		w.Address(), // creator
		gitopiatypes.RepositoryId{
			Id:   ownerId,
			Name: repoName,
		},
		title,
		description,
		labelIDs,
		0, // BountyId
		assignees,
		nil, // BountyAmount
		0,   // BountyExpiry
	)

	// Broadcast the transaction
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return "", err
	}

	var resp gitopiatypes.MsgCreateIssueResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return "", err
	}

	return strconv.FormatUint(resp.Iid, 10), nil
}

// CreatePullRequest signs & broadcasts a MsgCreatePullRequest.
func (c *Client) CreatePullRequest(
	ctx context.Context,
	w Wallet, // wallet able to SignAndBroadcast
	ownerId, repoName, title, description, headBranch, baseBranch string,
	assignees []string,
	labelIDs []uint64,
	issueIids []uint64,
) (int, error) {

	headRepoID := gitopiatypes.RepositoryId{Id: ownerId, Name: repoName}
	baseRepoID := gitopiatypes.RepositoryId{Id: ownerId, Name: repoName}

	msg := gitopiatypes.NewMsgCreatePullRequest(
		w.Address(), // creator
		title,
		description,
		headBranch,
		headRepoID,
		baseBranch,
		baseRepoID,
		nil, // reviewers
		assignees,
		labelIDs,
		issueIids,
	)

	// Broadcast the transaction
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return 0, err
	}

	var resp gitopiatypes.MsgCreatePullRequestResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return 0, err
	}

	return int(resp.Iid), nil
}

// CreateComment signs & broadcasts a MsgCreateComment on an issue.
func (c *Client) CreateComment(
	ctx context.Context,
	w Wallet,
	repoID uint64,
	issueIid uint64,
	body string,
) (*gitopiatypes.MsgCreateCommentResponse, error) {
	msg := gitopiatypes.NewMsgCreateComment(
		w.Address(), repoID, issueIid,
		gitopiatypes.CommentParentIssue,
		body, nil, "", "", 0,
	)
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	var resp gitopiatypes.MsgCreateCommentResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ToggleIssueState signs & broadcasts a MsgToggleIssueState.
func (c *Client) ToggleIssueState(
	ctx context.Context,
	w Wallet,
	repoID uint64,
	issueIid uint64,
	commentBody string,
) (*gitopiatypes.MsgToggleIssueStateResponse, error) {
	msg := gitopiatypes.NewMsgToggleIssueState(w.Address(), repoID, issueIid)
	msg.CommentBody = commentBody
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	var resp gitopiatypes.MsgToggleIssueStateResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddIssueLabels signs & broadcasts a MsgAddIssueLabels.
func (c *Client) AddIssueLabels(
	ctx context.Context,
	w Wallet,
	repoID uint64,
	issueIid uint64,
	labelIDs []uint64,
) (*gitopiatypes.MsgAddIssueLabelsResponse, error) {
	msg := gitopiatypes.NewMsgAddIssueLabels(w.Address(), repoID, issueIid, labelIDs)
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	var resp gitopiatypes.MsgAddIssueLabelsResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemoveIssueLabels signs & broadcasts a MsgRemoveIssueLabels.
func (c *Client) RemoveIssueLabels(
	ctx context.Context,
	w Wallet,
	repoID uint64,
	issueIid uint64,
	labelIDs []uint64,
) (*gitopiatypes.MsgRemoveIssueLabelsResponse, error) {
	msg := gitopiatypes.NewMsgRemoveIssueLabels(w.Address(), repoID, issueIid, labelIDs)
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	var resp gitopiatypes.MsgRemoveIssueLabelsResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddIssueAssignees signs & broadcasts a MsgAddIssueAssignees.
func (c *Client) AddIssueAssignees(
	ctx context.Context,
	w Wallet,
	repoID uint64,
	issueIid uint64,
	assignees []string,
) (*gitopiatypes.MsgAddIssueAssigneesResponse, error) {
	msg := gitopiatypes.NewMsgAddIssueAssignees(w.Address(), repoID, issueIid, assignees)
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	var resp gitopiatypes.MsgAddIssueAssigneesResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RemoveIssueAssignees signs & broadcasts a MsgRemoveIssueAssignees.
func (c *Client) RemoveIssueAssignees(
	ctx context.Context,
	w Wallet,
	repoID uint64,
	issueIid uint64,
	assignees []string,
) (*gitopiatypes.MsgRemoveIssueAssigneesResponse, error) {
	msg := gitopiatypes.NewMsgRemoveIssueAssignees(w.Address(), repoID, issueIid, assignees)
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	var resp gitopiatypes.MsgRemoveIssueAssigneesResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPullRequest returns a single pull request by owner, repo name, and pull IID.
func (c *Client) GetPullRequest(
	ctx context.Context,
	owner, repoName string,
	pullIid uint64,
) (*gitopiatypes.PullRequest, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.RepositoryPullRequest(ctx, &gitopiatypes.QueryGetRepositoryPullRequestRequest{
		Id:             owner,
		RepositoryName: repoName,
		PullIid:        pullIid,
	})
	if err != nil {
		return nil, err
	}
	return resp.PullRequest, nil
}

// CreatePullRequestComment signs & broadcasts a MsgCreateComment on a pull request.
func (c *Client) CreatePullRequestComment(
	ctx context.Context,
	w Wallet,
	repoID uint64,
	pullIid uint64,
	body, diffHunk, path string,
	position uint64,
) (*gitopiatypes.MsgCreateCommentResponse, error) {
	msg := gitopiatypes.NewMsgCreateComment(
		w.Address(), repoID, pullIid,
		gitopiatypes.CommentParentPullRequest,
		body, nil, diffHunk, path, position,
	)
	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}
	var resp gitopiatypes.MsgCreateCommentResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPullRequests lists all pull requests for a repository.
func (c *Client) ListPullRequests(ctx context.Context, owner, repoName string, limit uint64) ([]*gitopiatypes.PullRequest, error) {
	queryClient := gitopiatypes.NewQueryClient(c.conn)
	res, err := queryClient.RepositoryPullRequestAll(ctx, &gitopiatypes.QueryAllRepositoryPullRequestRequest{
		Id:             owner,
		RepositoryName: repoName,
		Pagination: &query.PageRequest{
			Limit: limit,
		},
	})
	if err != nil {
		return nil, err
	}
	return res.PullRequest, nil
}

// ForkRepository signs & broadcasts a MsgForkRepository.
func (c *Client) ForkRepository(
	ctx context.Context,
	w Wallet,
	owner, repoName, forkName, forkDescription, branch, forkOwner string,
) (*gitopiatypes.MsgForkRepositoryResponse, error) {
	msg := gitopiatypes.NewMsgForkRepository(
		w.Address(),
		gitopiatypes.RepositoryId{Id: owner, Name: repoName},
		forkName,
		forkDescription,
		branch,
		forkOwner,
	)

	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgForkRepositoryResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// ToggleRepositoryForking signs & broadcasts a MsgToggleRepositoryForking.
// It flips the AllowForking flag on the repository and returns the new state.
func (c *Client) ToggleRepositoryForking(
	ctx context.Context,
	w Wallet,
	owner, repoName string,
) (*gitopiatypes.MsgToggleRepositoryForkingResponse, error) {
	msg := &gitopiatypes.MsgToggleRepositoryForking{
		Creator:      w.Address(),
		RepositoryId: gitopiatypes.RepositoryId{Id: owner, Name: repoName},
	}

	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgToggleRepositoryForkingResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// InvokeMergePullRequest signs & broadcasts a MsgInvokeMergePullRequest.
// The v6 chain requires BaseCommitSha to match the current base branch SHA,
// so we query the base branch SHA before building the message.
func (c *Client) InvokeMergePullRequest(
	ctx context.Context,
	w Wallet,
	owner string,
	repoName string,
	repositoryID uint64,
	prNumber uint64,
	provider string,
) (string, error) {

	// Look up the PR to find the base branch, then get its SHA.
	q := gitopiatypes.NewQueryClient(c.conn)
	prResp, err := q.RepositoryPullRequest(ctx, &gitopiatypes.QueryGetRepositoryPullRequestRequest{
		Id:             owner,
		RepositoryName: repoName,
		PullIid:        prNumber,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get pull request: %w", err)
	}

	branchResp, err := q.RepositoryBranch(ctx, &gitopiatypes.QueryGetRepositoryBranchRequest{
		Id:             owner,
		RepositoryName: repoName,
		BranchName:     prResp.PullRequest.Base.Branch,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get base branch: %w", err)
	}
	baseCommitSha := branchResp.Branch.Sha

	msg := gitopiatypes.NewMsgInvokeMergePullRequest(
		w.Address(),  // creator
		repositoryID, // repositoryId
		prNumber,     // iid
		provider,     // git server provider
		baseCommitSha,
	)

	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return "", err
	}

	// The response for an invoke message might not contain data,
	// a successful transaction hash is sufficient confirmation.
	if txResponse != nil && txResponse.TxResponse != nil {
		return txResponse.TxResponse.TxHash, nil
	}

	return "", fmt.Errorf("transaction broadcast successful but no response received")
}

// GetBranch returns a single branch by name using a direct gRPC query.
// Avoids pagination issues with ListBranches on repos with many branches.
func (c *Client) GetBranch(
	ctx context.Context,
	ownerID, repoName, branchName string,
) (*gitopiatypes.Branch, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.RepositoryBranch(ctx, &gitopiatypes.QueryGetRepositoryBranchRequest{
		Id:             ownerID,
		RepositoryName: repoName,
		BranchName:     branchName,
	})
	if err != nil {
		return nil, err
	}
	return &resp.Branch, nil
}

// ListTags returns up to `limit` tags for a repository.
// Note: RepositoryTagAll returns []Tag (value type), not []*Tag.
func (c *Client) ListTags(
	ctx context.Context,
	ownerID, repoName string,
	limit uint64,
) ([]gitopiatypes.Tag, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.RepositoryTagAll(ctx, &gitopiatypes.QueryAllRepositoryTagRequest{
		Id:             ownerID,
		RepositoryName: repoName,
		Pagination:     &query.PageRequest{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	return resp.Tag, nil
}

// ListCommits fetches commit history via the git server HTTP gateway.
// The Gitopia SDK has no gRPC commit query; commits are served by the git server.
func (c *Client) ListCommits(
	ctx context.Context,
	repoID uint64,
	initCommitID string,
	limit int,
) ([]map[string]any, int, error) {
	_ = config.LoadGitConfig()
	url := fmt.Sprintf("%s/commits", config.GitServerHost)

	payload := map[string]any{
		"repository_id":  repoID,
		"init_commit_id": initCommitID,
		"pagination":     map[string]any{"count_total": true, "limit": limit},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal commits request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch commits from %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("bad status from %s: %s", url, resp.Status)
	}

	var result struct {
		Commits    []map[string]any `json:"commits"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to decode commits response: %w", err)
	}

	return result.Commits, result.Pagination.Total, nil
}

// DiffEntry represents a single file change in a pull request diff.
type DiffEntry struct {
	FileName string         `json:"file_name"`
	Stat     map[string]int `json:"stat"`
	Patch    string         `json:"patch"`
	Type     []string       `json:"type"`
	FileSize int            `json:"file_size"`
}

// GetPullRequestDiff fetches the diff between two commits via the git server HTTP gateway.
func (c *Client) GetPullRequestDiff(
	ctx context.Context,
	baseRepoID, headRepoID, baseCommitSHA, headCommitSHA string,
	limit int,
) ([]DiffEntry, error) {
	_ = config.LoadGitConfig()
	url := fmt.Sprintf("%s/pull/diff", config.GitServerHost)

	if limit <= 0 {
		limit = 50
	}

	payload := map[string]any{
		"base_repository_id": baseRepoID,
		"head_repository_id": headRepoID,
		"base_commit_sha":    baseCommitSHA,
		"head_commit_sha":    headCommitSHA,
		"pagination":         map[string]any{"offset": 0, "limit": limit},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal diff request: %w", err)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch diff from %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from %s: %s", url, resp.Status)
	}

	var result struct {
		Diff []DiffEntry `json:"diff"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode diff response: %w", err)
	}

	return result.Diff, nil
}

// ListReleases returns up to `limit` releases for a repository.
func (c *Client) ListReleases(
	ctx context.Context,
	ownerID, repoName string,
	limit uint64,
) ([]*gitopiatypes.Release, error) {
	q := gitopiatypes.NewQueryClient(c.conn)
	resp, err := q.RepositoryReleaseAll(ctx, &gitopiatypes.QueryAllRepositoryReleaseRequest{
		Id:             ownerID,
		RepositoryName: repoName,
		Pagination:     &query.PageRequest{Limit: limit},
	})
	if err != nil {
		return nil, err
	}
	return resp.Release, nil
}

// CreateRelease signs & broadcasts a MsgCreateRelease.
func (c *Client) CreateRelease(
	ctx context.Context,
	w Wallet,
	ownerID, repoName string,
	tagName, target, name, description string,
	draft, preRelease bool,
	provider string,
) (*gitopiatypes.MsgCreateReleaseResponse, error) {
	msg := gitopiatypes.NewMsgCreateRelease(
		w.Address(),
		gitopiatypes.RepositoryId{Id: ownerID, Name: repoName},
		tagName,
		target,
		name,
		description,
		"",    // attachments (omitted for v0.1.0)
		draft,
		preRelease,
		false, // isTag
		provider,
	)

	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgCreateReleaseResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreateRepositoryLabel signs & broadcasts a MsgCreateRepositoryLabel.
func (c *Client) CreateRepositoryLabel(
	ctx context.Context,
	w Wallet,
	ownerID, repoName string,
	labelName, color, description string,
) (*gitopiatypes.MsgCreateRepositoryLabelResponse, error) {
	msg := gitopiatypes.NewMsgCreateRepositoryLabel(
		w.Address(),
		gitopiatypes.RepositoryId{Id: ownerID, Name: repoName},
		labelName,
		color,
		description,
	)

	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgCreateRepositoryLabelResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteRepositoryLabel signs & broadcasts a MsgDeleteRepositoryLabel.
func (c *Client) DeleteRepositoryLabel(
	ctx context.Context,
	w Wallet,
	ownerID, repoName string,
	labelID uint64,
) (*gitopiatypes.MsgDeleteRepositoryLabelResponse, error) {
	msg := gitopiatypes.NewMsgDeleteRepositoryLabel(
		w.Address(),
		gitopiatypes.RepositoryId{Id: ownerID, Name: repoName},
		labelID,
	)

	txResponse, err := w.SignAndBroadcast(ctx, c.conn, []sdk.Msg{msg})
	if err != nil {
		return nil, err
	}

	var resp gitopiatypes.MsgDeleteRepositoryLabelResponse
	if err := signing.ParseTxResponse(signing.NewMarshaler(), txResponse, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
