# Prompt Cookbook

Practical prompt recipes for common workflows with the Gitopia MCP Server.

Each recipe lists the **goal**, **prerequisite tools**, and **step-by-step prompts** with expected output.

---

## Bounty Hunting

### Discover and Evaluate Bounties

**Goal:** Find open bounties, evaluate their complexity, and pick the best one to work on.

**Tools:** `list_bounties`, `get_bounty`, `get_issue`

```
Step 1: List all available bounties
→ list_bounties (limit: 50)

Step 2: Get details on a promising bounty
→ get_bounty (bounty_id: <id from step 1>)

Step 3: Read the linked issue for context
→ get_issue (owner: <repo_owner>, name: <repo_name>, issue_iid: <parent_iid>)

Step 4: Evaluate complexity vs reward
→ Compare bounty amount against issue requirements
```

### Claim and Fix a Bounty

**Goal:** Fix the issue linked to a bounty and submit a PR.

**Tools:** `git_clone`, `git_push`, `create_pull_request`, `comment_on_issue` + your editor's built-in file and git tools

```
Step 1: Clone the target repository
→ git_clone (repo_url: "gitopia://<owner>/<repo>", local_path: "<repo>")

Step 2: Create a feature branch
→ Use your built-in git tools to create and checkout a branch

Step 3: Investigate and implement the fix
→ Use your built-in file search, read, and edit tools

Step 4: Commit and push changes
→ Stage and commit with your built-in git tools, then git_push

Step 5: Create a pull request
→ create_pull_request (owner, name, title, description, head_branch, base_branch, issue_iids)

Step 6: Comment on the issue
→ comment_on_issue (owner, name, issue_iid, body: "Submitted PR")
```

---

## Code Review

### Review a Pull Request

**Goal:** Review code changes in a PR and leave comments.

**Tools:** `get_pull_request`, `get_pull_request_diff`, `get_file_contents`, `comment_on_pull_request`

```
Step 1: Get PR details
→ get_pull_request (owner, name, pull_iid)

Step 2: Get the code diff
→ get_pull_request_diff (owner, name, pull_iid)

Step 3: Review the changed files
→ Examine the diff patches. Use get_file_contents for additional context on specific files.

Step 4: Leave review comments
→ comment_on_pull_request (owner, name, pull_iid, body: "<feedback>")
→ For inline comments: include diff_hunk, path, and position

Step 5: Summarize the review
→ comment_on_pull_request with overall assessment
```

### List and Triage PRs

**Goal:** Review all open PRs for a repository.

**Tools:** `list_pull_requests`, `get_pull_request`

```
Step 1: List all pull requests
→ list_pull_requests (owner, name)

Step 2: Get details on each open PR
→ get_pull_request for each PR of interest

Step 3: Prioritize by age, size, and reviewer assignment
```

---

## Project Setup

### Bootstrap a New Repository

**Goal:** Create a new project with initial structure.

**Tools:** `bootstrap_repo`

```
Step 1: Create and initialize the repository
→ bootstrap_repo (
    owner_id: "<your_username>",
    name: "my-project",
    description: "A new project",
    create_readme: true,
    create_gitignore: true,
    gitignore_template: "go",
    initial_branch: "main"
  )

Expected output: Repository URL, local path, commit hash
```

### Configure a DAO

**Goal:** Set up a DAO for collaborative development.

**Tools:** `create_dao`, `get_dao`, `set_active_dao`

```
Step 1: Create the DAO
→ create_dao (
    name: "my-team",
    description: "Team DAO for collaborative development",
    voting_period: "2",
    percentage: "0.50"
  )

Step 1b: Look up the DAO's group_id and group_policy_address
→ get_dao (id: "my-team")

Step 2: Switch to DAO context
→ set_active_dao (dao_name: "my-team")

Step 3: Create repositories under the DAO
→ bootstrap_repo (owner_id: "<dao_address>", name: "team-project", ...)
```

---

## Issue Management

### Create and Assign Issues

**Goal:** Create issues and manage their lifecycle.

**Tools:** `create_issue`, `update_issue`, `comment_on_issue`

```
Step 1: Create an issue
→ create_issue (owner, name, title: "Bug: ...", description: "## Steps to reproduce\n...")

Step 2: Assign someone
→ update_issue (owner, name, issue_iid, add_assignees: ["username"])

Step 3: Add labels
→ update_issue (owner, name, issue_iid, add_labels: [1, 2])

Step 4: Comment with analysis
→ comment_on_issue (owner, name, issue_iid, body: "Initial analysis: ...")

Step 5: Close the issue when resolved
→ update_issue (owner, name, issue_iid, toggle_state: true, state_comment: "Fixed in PR #X")
```

### Triage Issues

**Goal:** Review and categorize open issues.

**Tools:** `list_issues`, `get_issue`, `update_issue`

```
Step 1: List all issues
→ list_issues (owner, name)

Step 2: Get details on each issue
→ get_issue for issues that need triage

Step 3: Categorize with labels
→ update_issue with add_labels for bug/feature/enhancement

Step 4: Assign to team members
→ update_issue with add_assignees
```

---

## Contribution Workflow

### Fork, Fix, and PR

**Goal:** Contribute to someone else's repository.

**Tools:** `fork_repository`, `git_clone`, `create_feature_branch_pr`

```
Step 1: Fork the repository
→ fork_repository (owner: "upstream-owner", name: "project")

Step 2: Clone your fork
→ git_clone (repo_url: "gitopia://<your_username>/project", local_path: "project")

Step 3: Create branch, make changes, and open PR
→ create_feature_branch_pr (
    repo_path: "project",
    owner: "<your_username>",
    name: "project",
    branch_name: "feature/my-change",
    base_branch: "main",
    files: [...],
    commit_message: "Add feature X",
    pr_title: "Add feature X",
    pr_description: "This PR adds ..."
  )
```

### Update an Existing PR

**Goal:** Push additional changes to an open PR.

**Tools:** `update_feature_branch`

```
Step 1: Update the feature branch
→ update_feature_branch (
    repo_path: "project",
    branch_name: "feature/my-change",
    files: [...],
    commit_message: "Address review feedback"
  )
```

---

## Using MCP Prompts

The server provides built-in prompt templates for common workflows:

### fix-issue Prompt
```
Prompt: fix-issue
Arguments: owner="myorg", repo="myproject", issue_number="42"
```
Generates a complete step-by-step workflow to fix issue #42.

### review-pr Prompt
```
Prompt: review-pr
Arguments: owner="myorg", repo="myproject", pull_iid="15"
```
Generates a code review workflow for PR #15.

### hunt-bounty Prompt
```
Prompt: hunt-bounty
Arguments: min_amount="1000" (optional), state="active" (optional)
```
Generates a bounty discovery and evaluation workflow.

---

## Using Dry-Run Mode

Set `DRY_RUN=true` environment variable or `"dry_run": true` in config to preview chain transactions without broadcasting.

```
# Preview what create_issue would do
DRY_RUN=true
→ create_issue (owner, name, title, description)
→ Returns JSON preview of the transaction parameters
```

This is useful for testing agent workflows before committing real transactions to the blockchain.

---

## Using MCP Resources

Browse Gitopia data through URI templates:

```
# Read a repository
Resource URI: gitopia://repos/myorg/myproject

# Read an issue
Resource URI: gitopia://repos/myorg/myproject/issues/42

# Read a pull request
Resource URI: gitopia://repos/myorg/myproject/pulls/15

# Read a bounty
Resource URI: gitopia://bounties/123
```

Resources return JSON data and can be used by MCP clients for browsable data access.
