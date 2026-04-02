package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gitopia/gitopia-mcp-server/internal/gitopia"
	"github.com/gitopia/gitopia-mcp-server/internal/logging"
	gitopiatypes "github.com/gitopia/gitopia/v6/x/gitopia/types"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ---- Gitopia API Handlers ----

type ListReposParams struct {
	Owner string `json:"owner" jsonschema:"Gitopia account name"`
}

func (h *ToolHandler) ListRepos(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params ListReposParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	repos, err := h.GClient.ListUserRepos(ctx, params.Owner, 50)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list repos for '%s': %s", params.Owner, err)
	}
	out := make([]map[string]any, 0, len(repos))
	for _, r := range repos {
		out = append(out, map[string]any{
			"name":        r.Name,
			"id":          r.Id,
			"description": r.Description,
			"stars":       r.Stargazers,
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d repositories for '%s'", len(out), params.Owner), out)
}

type CreateRepoParams struct {
	OwnerId     string `json:"owner_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

func (h *ToolHandler) CreateRepository(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateRepoParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_repo", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}
	ownerID := p.OwnerId
	if ownerID == "" {
		// Fall back to current user context (personal account or active DAO)
		contextMgr := h.getSessionContext(req)
		ownerID = contextMgr.GetOwnerID()
		if ownerID == "" {
			ownerID = w.Address()
		}
	}

	start := time.Now()
	resp, err := h.GClient.CreateRepository(ctx, w, ownerID, p.Name, p.Description)
	AuditLog("create_repo", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create repository: %s", err)
	}
	url := fmt.Sprintf("https://gitopia.com/%s/%s", resp.RepositoryId.Id, resp.RepositoryId.Name)
	return toolSuccess("Created repository", map[string]any{"url": url})
}

type GetRepoParams struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (h *ToolHandler) GetRepo(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params GetRepoParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	repo, err := h.GClient.GetRepository(ctx, params.Owner, params.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repo '%s/%s': %s", params.Owner, params.Name, err)
	}
	out := map[string]any{
		"name":        repo.Name,
		"id":          repo.Id,
		"description": repo.Description,
		"stars":       repo.Stargazers,
		"forks":       repo.Forks,
		"owner":       repo.Owner.Id,
	}
	return toolSuccessData(fmt.Sprintf("Repository '%s/%s'", params.Owner, params.Name), out)
}

type ListBranchesParams struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (h *ToolHandler) ListBranches(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params ListBranchesParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	branches, err := h.GClient.ListBranches(ctx, params.Owner, params.Name, 100)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list branches for '%s/%s': %s", params.Owner, params.Name, err)
	}
	out := make([]map[string]any, 0, len(branches))
	for _, b := range branches {
		out = append(out, map[string]any{"name": b.Name, "sha": b.Sha})
	}
	return toolSuccessData(fmt.Sprintf("Found %d branches for '%s/%s'", len(out), params.Owner, params.Name), out)
}

type GetFileParams struct {
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

func (h *ToolHandler) GetFile(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params GetFileParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	content, err := h.GClient.GetFile(ctx, params.Owner, params.Name, params.Branch, params.Path)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get file '%s' from '%s/%s@%s': %s", params.Path, params.Owner, params.Name, params.Branch, err)
	}
	// Exception: raw content delivery, no envelope
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(content)}},
	}, nil, nil
}

type ListIssuesParams struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (h *ToolHandler) ListIssues(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params ListIssuesParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	issues, err := h.GClient.ListIssues(ctx, params.Owner, params.Name, 100)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list issues for '%s/%s': %s", params.Owner, params.Name, err)
	}
	out := make([]map[string]any, 0, len(issues))
	for _, i := range issues {
		out = append(out, map[string]any{
			"number":      i.Iid,
			"title":       i.Title,
			"state":       i.State.String(),
			"author":      i.Creator,
			"description": i.Description,
			"comments":    i.CommentsCount,
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d issues for '%s/%s'", len(out), params.Owner, params.Name), out)
}

type GetIssueParams struct {
	Owner    string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name     string `json:"name" jsonschema:"Repository name"`
	IssueIid uint64 `json:"issue_iid" jsonschema:"Issue number (IID)"`
}

func (h *ToolHandler) GetIssue(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params GetIssueParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	issue, err := h.GClient.GetIssue(ctx, params.Owner, params.Name, params.IssueIid)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get issue #%d from '%s/%s': %s", params.IssueIid, params.Owner, params.Name, err)
	}
	out := map[string]any{
		"number":      issue.Iid,
		"title":       issue.Title,
		"state":       issue.State.String(),
		"author":      issue.Creator,
		"description": issue.Description,
		"comments":    issue.CommentsCount,
		"labels":      issue.Labels,
		"assignees":   issue.Assignees,
		"bounties":    issue.Bounties,
		"created_at":  issue.CreatedAt,
		"updated_at":  issue.UpdatedAt,
		"closed_at":   issue.ClosedAt,
		"closed_by":   issue.ClosedBy,
	}
	return toolSuccessData(fmt.Sprintf("Issue #%d from '%s/%s'", params.IssueIid, params.Owner, params.Name), out)
}

type CreateIssueParams struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (h *ToolHandler) CreateIssue(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreateIssueParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_issue", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}
	start := time.Now()
	issueId, err := h.GClient.CreateIssue(ctx, w, p.Owner, p.Name, p.Title, p.Description, nil, nil)
	AuditLog("create_issue", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create issue: %s", err)
	}
	logging.Infof("Successfully created issue %s", issueId)
	url := fmt.Sprintf("https://gitopia.com/%s/%s/issues/%s", p.Owner, p.Name, issueId)
	return toolSuccess("Created issue", map[string]any{"issue_iid": issueId, "url": url})
}

type CommentOnIssueParams struct {
	Owner    string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name     string `json:"name" jsonschema:"Repository name"`
	IssueIid uint64 `json:"issue_iid" jsonschema:"Issue number (IID)"`
	Body     string `json:"body" jsonschema:"Comment body text"`
}

func (h *ToolHandler) CommentOnIssue(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CommentOnIssueParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("comment_on_issue", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	repo, err := h.GClient.GetRepository(ctx, p.Owner, p.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s': %s", p.Owner, p.Name, err)
	}

	start := time.Now()
	_, err = h.GClient.CreateComment(ctx, w, repo.Id, p.IssueIid, p.Body)
	AuditLog("comment_on_issue", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to comment on issue #%d: %s", p.IssueIid, err)
	}

	return toolSuccess(fmt.Sprintf("Commented on issue #%d in %s/%s", p.IssueIid, p.Owner, p.Name), nil)
}

type UpdateIssueParams struct {
	Owner            string   `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name             string   `json:"name" jsonschema:"Repository name"`
	IssueIid         uint64   `json:"issue_iid" jsonschema:"Issue number (IID)"`
	ToggleState      bool     `json:"toggle_state,omitempty" jsonschema:"Toggle issue open/closed state"`
	StateComment     string   `json:"state_comment,omitempty" jsonschema:"Comment to add when toggling state"`
	AddLabels        []uint64 `json:"add_labels,omitempty" jsonschema:"Label IDs to add"`
	RemoveLabels     []uint64 `json:"remove_labels,omitempty" jsonschema:"Label IDs to remove"`
	AddAssignees     []string `json:"add_assignees,omitempty" jsonschema:"Usernames to assign"`
	RemoveAssignees  []string `json:"remove_assignees,omitempty" jsonschema:"Usernames to unassign"`
}

func (h *ToolHandler) UpdateIssue(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p UpdateIssueParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("update_issue", p)
	}
	// Validate early: at least one action must be specified
	if !p.ToggleState && len(p.AddLabels) == 0 && len(p.RemoveLabels) == 0 && len(p.AddAssignees) == 0 && len(p.RemoveAssignees) == 0 {
		return toolErrorf(ErrValidation, "No update actions specified. Provide at least one of: toggle_state, add_labels, remove_labels, add_assignees, remove_assignees.")
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	repo, err := h.GClient.GetRepository(ctx, p.Owner, p.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s': %s", p.Owner, p.Name, err)
	}

	// Collect all update messages into a single atomic transaction
	var msgs []sdk.Msg
	var actions []string

	if p.ToggleState {
		msg := gitopiatypes.NewMsgToggleIssueState(w.Address(), repo.Id, p.IssueIid)
		msg.CommentBody = p.StateComment
		msgs = append(msgs, msg)
		actions = append(actions, "toggled state")
	}

	if len(p.AddLabels) > 0 {
		msgs = append(msgs, gitopiatypes.NewMsgAddIssueLabels(w.Address(), repo.Id, p.IssueIid, p.AddLabels))
		actions = append(actions, fmt.Sprintf("added %d labels", len(p.AddLabels)))
	}

	if len(p.RemoveLabels) > 0 {
		msgs = append(msgs, gitopiatypes.NewMsgRemoveIssueLabels(w.Address(), repo.Id, p.IssueIid, p.RemoveLabels))
		actions = append(actions, fmt.Sprintf("removed %d labels", len(p.RemoveLabels)))
	}

	if len(p.AddAssignees) > 0 {
		msgs = append(msgs, gitopiatypes.NewMsgAddIssueAssignees(w.Address(), repo.Id, p.IssueIid, p.AddAssignees))
		actions = append(actions, fmt.Sprintf("added %d assignees", len(p.AddAssignees)))
	}

	if len(p.RemoveAssignees) > 0 {
		msgs = append(msgs, gitopiatypes.NewMsgRemoveIssueAssignees(w.Address(), repo.Id, p.IssueIid, p.RemoveAssignees))
		actions = append(actions, fmt.Sprintf("removed %d assignees", len(p.RemoveAssignees)))
	}

	if len(msgs) == 0 {
		return toolErrorf(ErrValidation, "No update actions specified. Provide at least one of: toggle_state, add_labels, remove_labels, add_assignees, remove_assignees.")
	}

	detail := fmt.Sprintf("issue #%d: %s", p.IssueIid, strings.Join(actions, ", "))
	result, _, err := h.signOrHold(ctx, req, w, msgs, "update_issue", detail)
	if result != nil || err != nil {
		return result, nil, err
	}

	summary := fmt.Sprintf("Updated issue #%d in %s/%s: %v", p.IssueIid, p.Owner, p.Name, actions)
	return toolSuccess(summary, nil)
}

type GetPullRequestParams struct {
	Owner   string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name    string `json:"name" jsonschema:"Repository name"`
	PullIid uint64 `json:"pull_iid" jsonschema:"Pull request number (IID)"`
}

func (h *ToolHandler) GetPullRequest(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params GetPullRequestParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	pr, err := h.GClient.GetPullRequest(ctx, params.Owner, params.Name, params.PullIid)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get pull request #%d from '%s/%s': %s", params.PullIid, params.Owner, params.Name, err)
	}
	out := map[string]any{
		"iid":          pr.Iid,
		"title":        pr.Title,
		"state":        pr.State.String(),
		"description":  pr.Description,
		"head":         pr.Head,
		"base":         pr.Base,
		"creator":      pr.Creator,
		"reviewers":    pr.Reviewers,
		"assignees":    pr.Assignees,
		"labels":       pr.Labels,
		"created_at":   pr.CreatedAt,
		"updated_at":   pr.UpdatedAt,
		"merged_by":    pr.MergedBy,
		"merged_at":    pr.MergedAt,
		"comments":     pr.CommentsCount,
		"draft":        pr.Draft,
	}
	return toolSuccessData(fmt.Sprintf("Pull request #%d from '%s/%s'", params.PullIid, params.Owner, params.Name), out)
}

type GetPullRequestDiffParams struct {
	Owner   string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name    string `json:"name" jsonschema:"Repository name"`
	PullIid uint64 `json:"pull_iid" jsonschema:"Pull request number (IID)"`
	Limit   int    `json:"limit,omitempty" jsonschema:"Maximum number of file diffs to return (default 50)"`
}

func (h *ToolHandler) GetPullRequestDiff(
	ctx context.Context,
	req *mcp.CallToolRequest,
	params GetPullRequestDiffParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	pr, err := h.GClient.GetPullRequest(ctx, params.Owner, params.Name, params.PullIid)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get pull request #%d from '%s/%s': %s", params.PullIid, params.Owner, params.Name, err)
	}
	if pr.Head == nil || pr.Base == nil {
		return toolErrorf(ErrValidation, "Pull request #%d has no head or base branch information", params.PullIid)
	}

	baseRepoID := fmt.Sprintf("%d", pr.Base.RepositoryId)
	headRepoID := fmt.Sprintf("%d", pr.Head.RepositoryId)
	baseCommitSHA := pr.Base.CommitSha
	headCommitSHA := pr.Head.CommitSha

	if baseCommitSHA == "" || headCommitSHA == "" {
		return toolErrorf(ErrValidation, "Pull request #%d is missing commit SHAs (base=%q, head=%q)", params.PullIid, baseCommitSHA, headCommitSHA)
	}

	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	diffs, err := h.GClient.GetPullRequestDiff(ctx, baseRepoID, headRepoID, baseCommitSHA, headCommitSHA, limit)
	if err != nil {
		return toolErrorf(ErrInternal, "Failed to fetch diff for PR #%d: %s", params.PullIid, err)
	}

	// Build a summary + full patches
	var totalAdditions, totalDeletions int
	files := make([]map[string]any, 0, len(diffs))
	for _, d := range diffs {
		totalAdditions += d.Stat["addition"]
		totalDeletions += d.Stat["deletion"]
		files = append(files, map[string]any{
			"file_name":  d.FileName,
			"type":       d.Type,
			"additions":  d.Stat["addition"],
			"deletions":  d.Stat["deletion"],
			"patch":      d.Patch,
		})
	}

	out := map[string]any{
		"pull_iid":        pr.Iid,
		"head_branch":     pr.Head.Branch,
		"base_branch":     pr.Base.Branch,
		"total_files":     len(diffs),
		"total_additions": totalAdditions,
		"total_deletions": totalDeletions,
		"files":           files,
	}
	return toolSuccessData(fmt.Sprintf("Diff for PR #%d: %d files changed, +%d -%d", params.PullIid, len(diffs), totalAdditions, totalDeletions), out)
}

type CommentOnPullRequestParams struct {
	Owner    string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name     string `json:"name" jsonschema:"Repository name"`
	PullIid  uint64 `json:"pull_iid" jsonschema:"Pull request number (IID)"`
	Body     string `json:"body" jsonschema:"Comment body text"`
	DiffHunk string `json:"diff_hunk,omitempty" jsonschema:"Diff hunk for inline comment"`
	Path     string `json:"path,omitempty" jsonschema:"File path for inline comment"`
	Position uint64 `json:"position,omitempty" jsonschema:"Line position for inline comment"`
}

func (h *ToolHandler) CommentOnPullRequest(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CommentOnPullRequestParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("comment_on_pull_request", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s. Ensure GITOPIA_MNEMONIC is set.", err)
	}

	repo, err := h.GClient.GetRepository(ctx, p.Owner, p.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s': %s", p.Owner, p.Name, err)
	}

	start := time.Now()
	_, err = h.GClient.CreatePullRequestComment(ctx, w, repo.Id, p.PullIid, p.Body, p.DiffHunk, p.Path, p.Position)
	AuditLog("comment_on_pull_request", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to comment on pull request #%d: %s", p.PullIid, err)
	}

	return toolSuccess(fmt.Sprintf("Commented on pull request #%d in %s/%s", p.PullIid, p.Owner, p.Name), nil)
}

type ListPullRequestsParams struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Limit uint64 `json:"limit,omitempty"`
}

func (h *ToolHandler) ListPullRequests(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ListPullRequestsParams,
) (*mcp.CallToolResult, any, error) {
	if h.GClient == nil {
		return toolErrorf(ErrClientUnavailable, "Gitopia client is not available. Check server configuration.")
	}
	limit := p.Limit
	if limit == 0 {
		limit = 100
	}
	prs, err := h.GClient.ListPullRequests(ctx, p.Owner, p.Name, limit)
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to list pull requests for '%s/%s': %s", p.Owner, p.Name, err)
	}
	out := make([]map[string]any, 0, len(prs))
	for _, pr := range prs {
		out = append(out, map[string]any{
			"number":      pr.Iid,
			"title":       pr.Title,
			"state":       pr.State,
			"author":      pr.Creator,
			"description": pr.Description,
			"head":        pr.Head.Branch,
			"base":        pr.Base.Branch,
		})
	}
	return toolSuccessData(fmt.Sprintf("Found %d pull requests for '%s/%s'", len(out), p.Owner, p.Name), out)
}

type MergePullRequestParams struct {
	Owner    string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name     string `json:"name" jsonschema:"Repository name"`
	PullIid uint64 `json:"pull_iid" jsonschema:"Pull request number (IID)"`
	Provider string `json:"provider,omitempty" jsonschema:"Git server provider address (defaults to gitopia15nv5vf6fmww8cxr6emrzxjvj36x5n8xvsxsqpw)"`
}

func (h *ToolHandler) MergePullRequest(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p MergePullRequestParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("merge_pull_request", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}

	repo, err := h.GClient.GetRepository(ctx, p.Owner, p.Name)
	if err != nil {
		return toolErrorf(ErrNotFound, "Failed to get repository '%s/%s': %s", p.Owner, p.Name, err)
	}

	provider := p.Provider
	if provider == "" {
		provider = "gitopia15nv5vf6fmww8cxr6emrzxjvj36x5n8xvsxsqpw"
	}

	start := time.Now()
	txHash, err := h.GClient.InvokeMergePullRequest(ctx, w, p.Owner, p.Name, repo.Id, p.PullIid, provider)
	AuditLog("merge_pull_request", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to merge pull request #%d: %s", p.PullIid, err)
	}
	logging.Infof("Successfully invoked merge for PR #%d (Tx: %s)", p.PullIid, txHash[:10])
	return toolSuccess(fmt.Sprintf("Merged pull request #%d", p.PullIid), map[string]any{"tx_hash": txHash})
}

type CreatePullRequestParams struct {
	Owner       string   `json:"owner"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	HeadBranch  string   `json:"head_branch"`
	BaseBranch  string   `json:"base_branch"`
	Assignees   []string `json:"assignees,omitempty"`
	Labels      []uint64 `json:"labels,omitempty"`
	IssueIids   []uint64 `json:"issue_iids,omitempty"`
}

func (h *ToolHandler) CreatePullRequest(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p CreatePullRequestParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("create_pull_request", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}
	start := time.Now()
	prNumber, err := retryOnSequenceMismatch(func() (int, error) {
		return h.GClient.CreatePullRequest(ctx, w, p.Owner, p.Name, p.Title, p.Description, p.HeadBranch, p.BaseBranch, p.Assignees, p.Labels, p.IssueIids)
	})
	AuditLog("create_pull_request", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to create pull request: %s", err)
	}
	url := fmt.Sprintf("https://gitopia.com/%s/%s/pulls/%d", p.Owner, p.Name, prNumber)
	return toolSuccess("Created pull request", map[string]any{"pull_iid": prNumber, "url": url})
}

type ForkRepositoryParams struct {
	Owner           string `json:"owner" jsonschema:"Source repository owner (username or DAO name)"`
	Name            string `json:"name" jsonschema:"Source repository name"`
	ForkName        string `json:"fork_name,omitempty" jsonschema:"Name for the forked repository (defaults to source name)"`
	ForkDescription string `json:"fork_description,omitempty" jsonschema:"Description for the fork"`
	Branch          string `json:"branch,omitempty" jsonschema:"Branch to fork (defaults to all branches)"`
	ForkOwner       string `json:"fork_owner,omitempty" jsonschema:"Owner of the fork (defaults to authenticated user)"`
}

func (h *ToolHandler) ForkRepository(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ForkRepositoryParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("fork_repository", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}

	forkName := p.ForkName
	if forkName == "" {
		forkName = p.Name
	}

	forkOwner := p.ForkOwner
	if forkOwner == "" {
		forkOwner = w.Address()
	}

	start := time.Now()
	resp, err := retryOnSequenceMismatch(func() (int, error) {
		r, err := h.GClient.ForkRepository(ctx, w, p.Owner, p.Name, forkName, p.ForkDescription, p.Branch, forkOwner)
		if err != nil {
			return 0, err
		}
		return int(r.Id), nil
	})
	AuditLog("fork_repository", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to fork repository '%s/%s': %s", p.Owner, p.Name, err)
	}

	return toolSuccess(fmt.Sprintf("Forked repository %s/%s", p.Owner, p.Name), map[string]any{
		"fork_id":    resp,
		"fork_owner": forkOwner,
		"fork_name":  forkName,
	})
}

type ToggleRepositoryForkingParams struct {
	Owner string `json:"owner" jsonschema:"Repository owner (username or DAO name)"`
	Name  string `json:"name" jsonschema:"Repository name"`
}

func (h *ToolHandler) ToggleRepositoryForking(
	ctx context.Context,
	req *mcp.CallToolRequest,
	p ToggleRepositoryForkingParams,
) (*mcp.CallToolResult, any, error) {
	if h.ChainLimiter != nil {
		if err := h.ChainLimiter.Allow(); err != nil {
			return toolErrorf(ErrRateLimited, "%s", err)
		}
	}
	if h.DryRunMode {
		return DryRunResult("toggle_repository_forking", p)
	}
	w, err := gitopia.WalletFromHeaders(req.Session, h.GClient)
	if err != nil {
		return toolErrorf(ErrAuthFailed, "Authentication failed: %s", err)
	}

	start := time.Now()
	resp, err := retryOnSequenceMismatch(func() (int, error) {
		r, err := h.GClient.ToggleRepositoryForking(ctx, w, p.Owner, p.Name)
		if err != nil {
			return 0, err
		}
		if r.AllowForking {
			return 1, nil
		}
		return 0, nil
	})
	AuditLog("toggle_repository_forking", w.Address(), err == nil, time.Since(start), "")
	if err != nil {
		return toolErrorf(ErrChainTx, "Failed to toggle repository forking for '%s/%s': %s", p.Owner, p.Name, err)
	}

	allowForking := resp == 1
	return toolSuccess("Toggled repository forking", map[string]any{"allow_forking": allowForking})
}
