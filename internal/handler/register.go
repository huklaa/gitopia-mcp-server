package handler

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all tools, prompts, and resource templates on the
// given MCP server using handlers from h and the provided gitopia client.
// This is extracted from cmd/server/main.go so that both the production
// binary and integration tests share the exact same registration.
func RegisterAll(srv *mcp.Server, h *ToolHandler) {
	registerPrompts(srv, h)
	registerResources(srv, h)
	registerTools(srv, h)
}

func registerPrompts(srv *mcp.Server, h *ToolHandler) {
	ph := &PromptHandler{}
	srv.AddPrompt(&mcp.Prompt{
		Name:        "fix-issue",
		Description: "Complete workflow to fix an issue: clone, branch, fix, test, PR",
		Arguments: []*mcp.PromptArgument{
			{Name: "owner", Description: "Repository owner", Required: true},
			{Name: "repo", Description: "Repository name", Required: true},
			{Name: "issue_number", Description: "Issue IID to fix", Required: true},
		},
	}, ph.FixIssuePrompt)

	srv.AddPrompt(&mcp.Prompt{
		Name:        "review-pr",
		Description: "Complete workflow to review a pull request: inspect, comment, approve/request changes",
		Arguments: []*mcp.PromptArgument{
			{Name: "owner", Description: "Repository owner", Required: true},
			{Name: "repo", Description: "Repository name", Required: true},
			{Name: "pr_number", Description: "Pull request IID to review", Required: true},
		},
	}, ph.ReviewPRPrompt)

	srv.AddPrompt(&mcp.Prompt{
		Name:        "hunt-bounty",
		Description: "Discover and work on bounties: list, evaluate, claim, fix, submit",
		Arguments: []*mcp.PromptArgument{
			{Name: "min_amount", Description: "Minimum bounty amount to consider"},
			{Name: "state", Description: "Filter bounties by state"},
		},
	}, ph.HuntBountyPrompt)
}

func registerResources(srv *mcp.Server, h *ToolHandler) {
	rh := &ResourceHandler{GClient: h.GClient}
	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "gitopia://repos/{owner}/{name}",
		Name:        "gitopia-repository",
		Description: "Repository details including name, description, stars, forks",
		MIMEType:    "application/json",
	}, rh.HandleRepoResource)

	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "gitopia://repos/{owner}/{name}/issues/{iid}",
		Name:        "gitopia-issue",
		Description: "Issue details including title, state, labels, assignees",
		MIMEType:    "application/json",
	}, rh.HandleIssueResource)

	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "gitopia://repos/{owner}/{name}/pulls/{iid}",
		Name:        "gitopia-pull-request",
		Description: "Pull request details including title, state, head, base, reviewers",
		MIMEType:    "application/json",
	}, rh.HandlePullRequestResource)

	srv.AddResourceTemplate(&mcp.ResourceTemplate{
		URITemplate: "gitopia://bounties/{id}",
		Name:        "gitopia-bounty",
		Description: "Bounty details including amount, state, expiry, creator",
		MIMEType:    "application/json",
	}, rh.HandleBountyResource)
}

// addToolIfEnabled registers a tool only if its toolset is active.
func addToolIfEnabled[In any](srv *mcp.Server, h *ToolHandler, tool *mcp.Tool, handler mcp.ToolHandlerFor[In, any]) {
	if h.ActiveToolsets != nil && !IsToolEnabled(tool.Name, h.ActiveToolsets) {
		return
	}
	mcp.AddTool(srv, tool, handler)
}

func registerTools(srv *mcp.Server, h *ToolHandler) {
	// ---- Context Management Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_user_context",
		Description: "Use this when you need to discover the active identity at the start of a session. " +
			"Returns username, wallet address, active DAO (if any), and a list of DAOs the user belongs to. " +
			"Automatically initializes context from wallet if not already done. " +
			"See also: set_active_dao, refresh_user_context.",
	}, withTrust(h, "get_user_context", h.GetUserContext))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "set_active_dao",
		Description: "Use this when you need to switch operations to a DAO context. " +
			"When a DAO is active, repository and issue operations use the DAO as the owner. " +
			"Pass dao_name to switch; omit or pass empty string to revert to personal account. " +
			"The DAO must appear in your get_user_context list. " +
			"See also: get_user_context.",
	}, withTrust(h, "set_active_dao", h.SetActiveDAO))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "refresh_user_context",
		Description: "Use this when context may be stale after DAO membership changes, new repository creation, " +
			"or other state-changing operations. Requires get_user_context to have been called first. " +
			"See also: get_user_context.",
	}, withTrust(h, "refresh_user_context", h.RefreshUserContext))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "claim_fee_grant",
		Description: "Use this to claim a fee grant from the Gitopia faucet for a wallet address. " +
			"If address is omitted, uses the current wallet. Fee grants allow signing transactions " +
			"without holding tokens. Best-effort: returns status even on failure. " +
			"See also: get_user_context.",
	}, withTrust(h, "claim_fee_grant", h.ClaimFeeGrant))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_user",
		Description: "Use this to create an on-chain Gitopia user for the current wallet. " +
			"Required before creating repos or interacting with the chain. " +
			"Requires 'username' (3-39 chars, alphanumeric and hyphens). " +
			"Automatically called during wallet auto-generation, but can also be used standalone. " +
			"See also: get_user_context.",
	}, withTrust(h, "create_user", h.CreateUser))

	// ---- Repository Management Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_repos",
		Description: "Use this when you need to browse or discover repositories for a user or DAO. " +
			"Returns a JSON array of repository objects with name, id, description, and star count. " +
			"Requires 'owner' (username or DAO name). Returns up to 50 repos. " +
			"See also: get_repo for full details on a specific repo.",
	}, withTrust(h, "list_repos", h.ListRepos))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_repo",
		Description: "Use this when you need to create a remote repository on Gitopia without local initialization. " +
			"Signs and broadcasts an on-chain transaction. Returns the repository URL. " +
			"owner_id is optional; if omitted, uses the current user context (personal or active DAO). " +
			"Requires 'name'. Optional: 'description'. " +
			"Prefer bootstrap_repo if you also want to initialize locally, add files, and push an initial commit. " +
			"See also: bootstrap_repo, git_clone.",
	}, withTrust(h, "create_repo", h.CreateRepository))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_repo",
		Description: "Use this when you need detailed information about a specific repository (e.g. to get repo_id for merge). " +
			"Returns name, id, description, stars, forks, and owner. " +
			"Requires 'owner' and 'name'. " +
			"See also: list_repos, git_clone.",
	}, withTrust(h, "get_repo", h.GetRepo))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_branches",
		Description: "Use this when you need to discover branches before cloning or creating a PR. " +
			"Returns a JSON array of branch objects with name and SHA. " +
			"Requires 'owner' and 'name'. Returns up to 100 branches. " +
			"See also: get_repo, create_pull_request.",
	}, withTrust(h, "list_branches", h.ListBranches))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_file_contents",
		Description: "Use this when you need to read a file from a remote Gitopia repository without cloning it first. " +
			"Fetches via the git server gateway. " +
			"Requires 'owner', 'name', 'branch', and 'path'. Returns the raw file content as text. " +
			"See also: git_clone.",
	}, withTrust(h, "get_file_contents", h.GetFile))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_tags",
		Description: "Use this when you need to discover version tags for a repository. " +
			"Returns a JSON array of tag objects with name, sha, created_at, and updated_at. " +
			"Requires 'owner' and 'name'. Optional: 'limit' (default 100). " +
			"See also: list_branches, list_releases.",
	}, withTrust(h, "list_tags", h.ListTags))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_commits",
		Description: "Use this when you need to review commit history for a repository branch. " +
			"Returns a JSON array of commit objects with id, title, author, committer, message, and created_at. " +
			"Requires 'owner' and 'name'. Optional: 'branch' (defaults to the repo's default branch), 'limit' (default 50). " +
			"See also: list_branches, get_repo.",
	}, withTrust(h, "list_commits", h.ListCommits))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_releases",
		Description: "Use this when you need to discover published releases for a repository. " +
			"Returns a JSON array of release objects with id, tag_name, name, description, draft, pre_release, and created_at. " +
			"Requires 'owner' and 'name'. Optional: 'limit' (default 50). " +
			"See also: list_tags, create_release.",
	}, withTrust(h, "list_releases", h.ListReleases))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_release",
		Description: "Use this when you want to publish a release for a tagged version. " +
			"The tag must already exist on the remote (push it via git first). " +
			"Signs and broadcasts an on-chain transaction. Returns the release ID. " +
			"Requires 'owner', 'name', 'tag_name', and 'release_name'. " +
			"Optional: 'target' (branch, defaults to the repo's default branch), " +
			"'description', 'draft', 'pre_release', 'provider'. " +
			"See also: list_releases, list_tags.",
	}, withTrust(h, "create_release", h.CreateRelease))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_labels",
		Description: "Use this when you need to see available labels for a repository before applying them to issues. " +
			"Returns a JSON array of label objects with id, name, color, and description. " +
			"Requires 'owner' and 'name'. " +
			"See also: create_label, update_issue.",
	}, withTrust(h, "list_labels", h.ListLabels))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_label",
		Description: "Use this when you need to create a new label for organizing issues in a repository. " +
			"Signs and broadcasts an on-chain transaction. Returns the label ID. " +
			"Requires 'owner', 'name', 'label_name', and 'color' (hex code like 'FF0000'). " +
			"Optional: 'description'. " +
			"See also: list_labels, delete_label.",
	}, withTrust(h, "create_label", h.CreateLabel))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "delete_label",
		Description: "Use this when you need to permanently remove a label from a repository. " +
			"Signs and broadcasts an on-chain transaction. Irreversible. " +
			"Requires 'owner', 'name', and 'label_id'. " +
			"See also: list_labels, create_label.",
	}, withTrust(h, "delete_label", h.DeleteLabel))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_issues",
		Description: "Use this when you need to browse issues for a repository or find issues to work on. " +
			"Returns a JSON array with number, title, state, author, description, and comment count. " +
			"Requires 'owner' and 'name'. Returns up to 100 issues. " +
			"Use get_issue for full details on a specific issue. " +
			"See also: get_issue, create_issue.",
	}, withTrust(h, "list_issues", h.ListIssues))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_issue",
		Description: "Use this when you need full issue context before working on it (labels, assignees, bounties, linked PRs). " +
			"Returns number, title, state, author, description, comments count, labels, assignees, bounties, " +
			"created_at, updated_at, closed_at, and closed_by. " +
			"Requires 'owner', 'name', and 'issue_iid'. " +
			"See also: list_issues, comment_on_issue, update_issue.",
	}, withTrust(h, "get_issue", h.GetIssue))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_issue",
		Description: "Use this when you need to report a bug, request a feature, or create a task. " +
			"Signs and broadcasts a transaction on-chain. Returns the issue URL. " +
			"Requires 'owner', 'name', 'title', and 'description'. " +
			"See also: get_issue, comment_on_issue.",
	}, withTrust(h, "create_issue", h.CreateIssue))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "comment_on_issue",
		Description: "Use this when you need to communicate findings, ask questions, or provide status updates on an issue. " +
			"Signs and broadcasts a transaction on-chain. " +
			"Requires 'owner', 'name', 'issue_iid', and 'body'. " +
			"See also: get_issue, update_issue.",
	}, withTrust(h, "comment_on_issue", h.CommentOnIssue))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "update_issue",
		Description: "Use this when you need to change an issue's state (open/close), labels, or assignees. " +
			"All actions are batched into a single atomic on-chain transaction. " +
			"Requires 'owner', 'name', 'issue_iid'. " +
			"Optional: 'toggle_state' (open/close), 'state_comment', 'add_labels', 'remove_labels', " +
			"'add_assignees', 'remove_assignees'. At least one action must be specified. " +
			"See also: get_issue, comment_on_issue.",
	}, withTrust(h, "update_issue", h.UpdateIssue))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_pull_requests",
		Description: "Use this when you need to check existing PRs before creating a new one, or review PR status. " +
			"Returns a JSON array of PR objects with number, title, state, author, description, head, and base branch. " +
			"Requires 'owner' and 'name'. Optional: 'limit' (default 100). " +
			"See also: create_pull_request, merge_pull_request.",
	}, withTrust(h, "list_pull_requests", h.ListPullRequests))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_pull_request",
		Description: "Use this when you need full pull request context before reviewing or merging. " +
			"Returns iid, title, state, description, head, base, creator, reviewers, assignees, labels, " +
			"created_at, updated_at, merged_by, merged_at, comments count, and draft status. " +
			"Requires 'owner', 'name', and 'pull_iid'. " +
			"See also: list_pull_requests, comment_on_pull_request, merge_pull_request.",
	}, withTrust(h, "get_pull_request", h.GetPullRequest))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_pull_request_diff",
		Description: "Use this when you need to review the actual code changes in a pull request. " +
			"Returns the unified diff (patches) for all changed files, with addition/deletion stats. " +
			"Requires 'owner', 'name', and 'pull_iid'. Optional: 'limit' (max files, default 50). " +
			"See also: get_pull_request, comment_on_pull_request.",
	}, withTrust(h, "get_pull_request_diff", h.GetPullRequestDiff))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "comment_on_pull_request",
		Description: "Use this when you need to leave review comments on a pull request. " +
			"Supports both general comments and inline comments on specific code lines. " +
			"Signs and broadcasts a transaction on-chain. " +
			"Requires 'owner', 'name', 'pull_iid', and 'body'. " +
			"Optional: 'diff_hunk', 'path', 'position' for inline comments. " +
			"See also: get_pull_request, list_pull_requests.",
	}, withTrust(h, "comment_on_pull_request", h.CommentOnPullRequest))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "merge_pull_request",
		Description: "Use this when a PR has been reviewed and is ready to merge. " +
			"Merges via an on-chain transaction. Returns the transaction hash. " +
			"Requires 'owner', 'name', and 'pull_iid'. " +
			"See also: get_pull_request, list_pull_requests.",
	}, withTrust(h, "merge_pull_request", h.MergePullRequest))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_pull_request",
		Description: "Use this when you have pushed a branch and want to open a PR for review. " +
			"Creates a pull request on Gitopia. Automatically retries on account sequence mismatch. " +
			"Requires 'owner', 'name', 'title', 'description', 'head_branch', and 'base_branch'. " +
			"Optional: 'assignees' (usernames), 'labels' (label IDs), 'issue_iids' (issue numbers to link). " +
			"Prefer create_feature_branch_pr for the full branch-commit-push-PR cycle. " +
			"See also: create_feature_branch_pr, git_push.",
	}, withTrust(h, "create_pull_request", h.CreatePullRequest))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "fork_repository",
		Description: "Use this when you need to fork a repository for contribution (fork -> branch -> PR workflow). " +
			"Signs and broadcasts an on-chain transaction. Returns the fork ID. " +
			"Requires 'owner' and 'name'. " +
			"Optional: 'fork_name' (defaults to source name), 'fork_description', " +
			"'branch' (specific branch to fork), 'fork_owner' (defaults to authenticated user). " +
			"See also: git_clone, create_pull_request.",
	}, withTrust(h, "fork_repository", h.ForkRepository))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "toggle_repository_forking",
		Description: "Use this when you need to enable or disable forking on a repository. " +
			"Toggles the AllowForking flag. Must be called before forking a repo that has forking disabled. " +
			"Signs and broadcasts an on-chain transaction. Returns the new AllowForking state. " +
			"Requires 'owner' and 'name'. " +
			"See also: fork_repository, get_repo.",
	}, withTrust(h, "toggle_repository_forking", h.ToggleRepositoryForking))

	// ---- Git Operation Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "git_clone",
		Description: "Use this when you need a local copy of a Gitopia repository to read or modify code. " +
			"Clones into a new workspace directory. Fails if target already exists. " +
			"Requires 'repo_url' (gitopia://owner/repo or owner/repo) and 'local_path' (workspace-relative). " +
			"Optional: 'branch' for a specific branch, 'depth' for shallow clone. " +
			"See also: get_file_contents (for reading without cloning), bootstrap_repo (for new repos).",
	}, withTrust(h, "git_clone", h.GitClone))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_feature_branch",
		Description: "Use this when you need to start working on a new feature in an existing local repo. " +
			"Creates and checks out a new branch. " +
			"Requires 'repo_path' and 'branch_name'. Optional: 'base_branch' (defaults to current branch). " +
			"Prefer create_feature_branch_pr for the full branch-change-commit-push-PR cycle. " +
			"See also: create_feature_branch_pr.",
	}, withTrust(h, "create_feature_branch", h.CreateFeatureBranch))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "sync_with_remote",
		Description: "Use this when you need to update a local repo with upstream changes before starting work. " +
			"Fetches and integrates changes from the remote. " +
			"Requires 'repo_path'. Optional: 'branch', 'remote' (default 'origin'), " +
			"'strategy' ('merge' or 'rebase', default 'merge'). " +
			"See also: git_clone, git_push.",
	}, withTrust(h, "sync_with_remote", h.SyncWithRemote))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "git_push",
		Description: "Use this when you have local commits ready to publish to the remote. " +
			"Requires 'repo_path'. Optional: 'remote' (default 'origin'), 'branch' (default current), " +
			"'force', 'set_upstream', 'dry_run'. " +
			"Prefer commit_and_push_changes for a single-step stage+commit+push workflow. " +
			"See also: commit_and_push_changes, create_pull_request.",
	}, withTrust(h, "git_push", h.GitPush))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "commit_and_push_changes",
		Description: "Use this when you want to stage, commit, and push in one step (most common push workflow). " +
			"Requires 'repo_path' and 'commit_message'. " +
			"Optional: 'files' (specific files to stage; empty stages all), 'branch', 'create_branch'. " +
			"Automatically sets upstream tracking. " +
			"See also: git_push, create_pull_request.",
	}, withTrust(h, "commit_and_push_changes", h.CommitAndPushChanges))

	// ---- Workflow Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_feature_branch_pr",
		Description: "Use this when you want the complete contribution workflow in one operation: " +
			"create branch, apply file changes, commit, push, and open a PR. " +
			"This is the recommended tool for code contributions. " +
			"Requires the repository to already be cloned locally (use git_clone first). " +
			"base_branch defaults to 'main'. " +
			"Use update_feature_branch to add commits to an existing branch/PR instead. " +
			"See also: git_clone, update_feature_branch, create_pull_request.",
	}, withTrust(h, "create_feature_branch_pr", h.CreateFeatureBranchPR))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "update_feature_branch",
		Description: "Use this when you need to push additional changes to an existing feature branch or open PR. " +
			"Modifies the branch by adding a new commit with file changes and pushing. " +
			"Requires 'repo_path', 'branch_name', 'files', and 'commit_message'. " +
			"The branch must already exist locally. " +
			"See also: create_feature_branch_pr.",
	}, withTrust(h, "update_feature_branch", h.UpdateFeatureBranch))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "bootstrap_repo",
		Description: "Use this when you need to create a brand new project from scratch (remote + local + initial commit). " +
			"Creates a Gitopia repository, initializes it locally, adds initial files (README, .gitignore), " +
			"and pushes the first commit. " +
			"Requires 'owner_id' and 'name'. Optional: 'local_path' (defaults to repo name), " +
			"'description', 'create_readme' (default true), 'create_gitignore', " +
			"'gitignore_template' (go, python, node, rust, java, .net, generic), 'initial_branch' (default 'main'). " +
			"See also: create_repo.",
	}, withTrust(h, "bootstrap_repo", h.BootstrapRepo))

	// ---- DAO Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_dao",
		Description: "Use this when you need DAO details including group_id and group_policy_address " +
			"(required for governance tools like list_members, submit_proposal, vote). " +
			"Returns name, address, description, group_id, group_policy_address, and metadata. " +
			"Requires 'id' (DAO name or address). " +
			"See also: create_dao, dao_list_members, set_active_dao.",
	}, withTrust(h, "get_dao", h.GetDao))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_dao",
		Description: "Use this when you need to create a new DAO for collaborative development. " +
			"Signs and broadcasts an on-chain transaction. Returns the new DAO ID. " +
			"Requires 'name' and 'description'. Optional: 'avatar_url', 'location', 'website', " +
			"'members' (defaults to creator with weight 1), 'voting_period' (default '2' hours), " +
			"'percentage' (default '0.50' = 50%). " +
			"See also: set_active_dao.",
	}, withTrust(h, "create_dao", h.CreateDao))

	// ---- DAO Governance Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "dao_list_members",
		Description: "Use this when you need to see who belongs to a DAO and their voting weights. " +
			"Returns a JSON array of member objects with address, weight, metadata, and added_at. " +
			"Provide 'dao' (name or address) OR 'group_id'. Optional: 'limit' (default 100). " +
			"See also: dao_list_proposals, create_dao.",
	}, withTrust(h, "dao_list_members", h.ListDaoMembers))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "dao_list_proposals",
		Description: "Use this when you need to review pending or past governance proposals for a DAO. " +
			"Returns a JSON array of proposal objects with id, title, summary, status, proposers, and voting_period_end. " +
			"Provide 'dao' (name or address) OR 'group_policy_address'. Optional: 'limit' (default 50). " +
			"See also: dao_get_proposal, dao_vote.",
	}, withTrust(h, "dao_list_proposals", h.ListDaoProposals))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "dao_get_proposal",
		Description: "Use this when you need full details about a specific DAO governance proposal, including tally results. " +
			"Returns id, title, summary, status, proposers, metadata, final_tally_result, voting_period_end, and executor_result. " +
			"Requires 'proposal_id'. " +
			"See also: dao_list_proposals, dao_vote, dao_exec.",
	}, withTrust(h, "dao_get_proposal", h.GetDaoProposal))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "dao_update_members",
		Description: "Use this when you need to add, remove, or change voting weights of DAO members. " +
			"Signs and broadcasts an on-chain transaction. Set a member's weight to '0' to remove them. " +
			"Provide 'dao' (name or address) OR 'group_id', plus 'member_updates' (array of {address, weight, metadata}). " +
			"The caller must be the group admin. " +
			"See also: dao_list_members, create_dao.",
	}, withTrust(h, "dao_update_members", h.UpdateDaoMembers))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "dao_submit_proposal",
		Description: "Use this when you need to create a text governance proposal for DAO voting. " +
			"Signs and broadcasts an on-chain transaction. Returns the new proposal ID. " +
			"Provide 'dao' (name or address) OR 'group_policy_address', plus 'title'. " +
			"Optional: 'summary', 'metadata', 'exec_try'. " +
			"The proposer is automatically set to the current wallet address. " +
			"See also: dao_list_proposals, dao_vote, dao_exec.",
	}, withTrust(h, "dao_submit_proposal", h.SubmitDaoProposal))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "dao_vote",
		Description: "Use this when you need to cast a vote on an active DAO proposal. " +
			"Signs and broadcasts an on-chain transaction. " +
			"Requires 'proposal_id' and 'option' (yes, no, abstain, or no_with_veto). " +
			"Optional: 'metadata', 'exec_try' (try to execute after voting). " +
			"See also: dao_get_proposal, dao_exec.",
	}, withTrust(h, "dao_vote", h.VoteDaoProposal))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "dao_exec",
		Description: "Use this when a proposal has passed and needs to be executed on-chain. " +
			"Signs and broadcasts an on-chain transaction. " +
			"Requires 'proposal_id'. The proposal must have passed (tally succeeded). " +
			"See also: dao_get_proposal, dao_vote.",
	}, withTrust(h, "dao_exec", h.ExecDaoProposal))

	// ---- Bounty Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "create_bounty",
		Description: "Use this when you want to incentivize work on an issue by attaching a reward. " +
			"Signs and broadcasts an on-chain transaction. Returns the bounty ID. " +
			"Requires 'owner', 'name', 'issue_iid', and 'amount' (array of {denom, amount} coins, amount as string). " +
			"Optional: 'expiry' (unix timestamp). " +
			"See also: list_bounties, get_bounty, update_bounty.",
	}, withTrust(h, "create_bounty", h.CreateBounty))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "update_bounty",
		Description: "Use this when you need to extend a bounty's expiry. " +
			"Requires 'bounty_id'. Optional: 'expiry' (unix timestamp). " +
			"See also: get_bounty, close_bounty.",
	}, withTrust(h, "update_bounty", h.UpdateBounty))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "close_bounty",
		Description: "Use this when a bounty should no longer accept claims (e.g. issue resolved without bounty). " +
			"Requires 'bounty_id'. Only the bounty creator can close it. " +
			"See also: get_bounty, delete_bounty.",
	}, withTrust(h, "close_bounty", h.CloseBounty))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "delete_bounty",
		Description: "Use this when you need to permanently remove a bounty (irreversible). " +
			"Requires 'bounty_id'. Only the bounty creator can delete it. " +
			"Consider close_bounty first if you just want to deactivate it. " +
			"See also: close_bounty.",
	}, withTrust(h, "delete_bounty", h.DeleteBounty))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_bounties",
		Description: "Use this when you need to discover available bounties for contribution or review bounty status. " +
			"Returns a JSON array of bounty objects with id, amount, state, repository_id, parent_iid, expire_at, and creator. " +
			"Optional: 'limit' (default 50). " +
			"Use get_bounty to get full details on a specific bounty. " +
			"See also: get_bounty, create_bounty.",
	}, withTrust(h, "list_bounties", h.ListBounties))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "get_bounty",
		Description: "Use this when you need full bounty details to evaluate whether to work on it. " +
			"Returns id, amount, state, repository_id, parent_iid, parent, expire_at, rewarded_to, " +
			"creator, created_at, updated_at. " +
			"Requires 'bounty_id'. " +
			"See also: list_bounties, get_issue.",
	}, withTrust(h, "get_bounty", h.GetBounty))

	// ---- Batch Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "batch_execute",
		Description: "Use this when you need to execute multiple operations in a single on-chain transaction for atomicity. " +
			"All operations succeed or fail together, saving gas and reducing round-trips. " +
			"Requires 'operations' (array of {tool, params}, max 10). " +
			"Supported tools: dao_vote, dao_update_members, comment_on_issue, comment_on_pull_request. " +
			"See also: dao_vote, dao_update_members.",
	}, withTrust(h, "batch_execute", h.BatchExecute))

	// ---- Approval Gate Tools ----
	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "confirm_transaction",
		Description: "Use this to broadcast a pending chain-write transaction that was held for approval. " +
			"Requires 'pending_id' (returned by the original tool call when APPROVAL_MODE is enabled). " +
			"The transaction must belong to the current session and not be expired. " +
			"See also: list_pending_transactions, reject_transaction.",
	}, withTrust(h, "confirm_transaction", h.ConfirmTransaction))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "reject_transaction",
		Description: "Use this to cancel a pending chain-write transaction without broadcasting. " +
			"Requires 'pending_id'. The transaction is permanently removed. " +
			"See also: list_pending_transactions, confirm_transaction.",
	}, withTrust(h, "reject_transaction", h.RejectTransaction))

	addToolIfEnabled(srv, h, &mcp.Tool{
		Name: "list_pending_transactions",
		Description: "Use this to see all pending chain-write transactions awaiting confirmation in the current session. " +
			"Returns pending_id, tool, detail, wallet, and expiry for each. " +
			"See also: confirm_transaction, reject_transaction.",
	}, withTrust(h, "list_pending_transactions", h.ListPendingTransactions))
}
