package handler

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PromptHandler holds dependencies for MCP prompt handlers.
type PromptHandler struct{}

// FixIssuePrompt returns a multi-step workflow to fix an issue.
func (ph *PromptHandler) FixIssuePrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	owner := req.Params.Arguments["owner"]
	repo := req.Params.Arguments["repo"]
	issueNumber := req.Params.Arguments["issue_number"]

	if owner == "" || repo == "" || issueNumber == "" {
		return nil, fmt.Errorf("missing required arguments: owner, repo, and issue_number are required")
	}

	prompt := fmt.Sprintf(`You are a software engineer tasked with fixing an issue in a Gitopia repository.

## Issue Details
- Repository: %s/%s
- Issue: #%s

## Workflow Steps

1. **Get issue details**: Use get_issue with owner="%s", name="%s", issue_iid=%s to understand the problem.

2. **Clone the repository**: Use git_clone with repo_url="gitopia://%s/%s" and local_path="%s" to get a local copy.

3. **Create a feature branch**: Create and checkout a new branch "fix/issue-%s" using your built-in git tools.

4. **Investigate the code**: Use your built-in file search and read tools to understand the relevant codebase.

5. **Implement the fix**: Use your built-in file editing tools to make the necessary changes.

6. **Write tests**: Add or update tests to verify the fix works correctly.

7. **Commit and push**: Stage, commit with a descriptive message referencing the issue, then use git_push to push to the remote.

8. **Create a pull request**: Use create_pull_request with:
   - owner="%s"
   - name="%s"
   - title="Fix #%s: [brief description]"
   - description="Fixes #%s\n\n[detailed description of changes]"
   - head_branch="fix/issue-%s"
   - base_branch="main"
   - issue_iids=[%s]

9. **Comment on the issue**: Use comment_on_issue to note that a PR has been submitted.

Begin by fetching the issue details to understand what needs to be fixed.`,
		owner, repo, issueNumber,
		owner, repo, issueNumber,
		owner, repo, repo,
		issueNumber,
		owner, repo, issueNumber, issueNumber, issueNumber, issueNumber,
	)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Fix issue #%s in %s/%s", issueNumber, owner, repo),
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: prompt}},
		},
	}, nil
}

// ReviewPRPrompt returns a multi-step workflow to review a pull request.
func (ph *PromptHandler) ReviewPRPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	owner := req.Params.Arguments["owner"]
	repo := req.Params.Arguments["repo"]
	prNumber := req.Params.Arguments["pr_number"]

	if owner == "" || repo == "" || prNumber == "" {
		return nil, fmt.Errorf("missing required arguments: owner, repo, and pr_number are required")
	}

	prompt := fmt.Sprintf(`You are a code reviewer tasked with reviewing a pull request on Gitopia.

## Pull Request Details
- Repository: %s/%s
- PR: #%s

## Review Workflow

1. **Get PR details**: Use get_pull_request with owner="%s", name="%s", pull_iid=%s to understand the PR scope.

2. **Get the diff**: Use get_pull_request_diff with owner="%s", name="%s", pull_iid=%s to see all code changes.

3. **Review the changes**:
   - Examine each file diff for code quality, correctness, and security issues.
   - If you need more context, use get_file_contents to read full files from the repository.
   - Check for adherence to project conventions and test coverage.

4. **Leave review comments**: Use comment_on_pull_request to provide feedback:
   - For general feedback: set body with your overall review.
   - For inline comments: set body, diff_hunk, path, and position for specific code feedback.

5. **Summarize your review**: Leave a final comment with:
   - Overall assessment (approve / request changes)
   - Summary of findings
   - Suggested improvements

Focus on correctness, security, maintainability, and test coverage.`,
		owner, repo, prNumber,
		owner, repo, prNumber,
		owner, repo, prNumber,
	)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Review PR #%s in %s/%s", prNumber, owner, repo),
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: prompt}},
		},
	}, nil
}

// HuntBountyPrompt returns a multi-step workflow to find and work on bounties.
func (ph *PromptHandler) HuntBountyPrompt(_ context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	minAmount := req.Params.Arguments["min_amount"]
	state := req.Params.Arguments["state"]

	filterNote := ""
	if minAmount != "" {
		filterNote += fmt.Sprintf("\n- Minimum amount: %s", minAmount)
	}
	if state != "" {
		filterNote += fmt.Sprintf("\n- Filter by state: %s", state)
	}

	prompt := fmt.Sprintf(`You are a bounty hunter looking for bounties to work on in the Gitopia ecosystem.
%s
## Bounty Hunting Workflow

1. **List available bounties**: Use list_bounties to discover open bounties.

2. **Filter and evaluate**: Review the bounty list and filter by:
   - Amount (prioritize higher-value bounties)
   - State (focus on active/open bounties)
   - Expiry (avoid nearly-expired bounties)

3. **Get bounty details**: For promising bounties, use get_bounty to get full details including the linked issue.

4. **Evaluate complexity**: For the top candidates:
   - Use get_issue to read the linked issue
   - Assess the complexity vs. reward ratio
   - Check if you have the required skills

5. **Pick the best bounty**: Choose the bounty with the best complexity-to-reward ratio that matches your skills.

6. **Execute the fix**: Follow the fix-issue workflow:
   - Use git_clone to get a local copy
   - Create a feature branch using your built-in git tools
   - Implement the fix using your built-in file editing tools
   - Write tests
   - Push with git_push, then use create_pull_request linking the issue

7. **Comment on the issue**: Use comment_on_issue to note that you're working on the bounty.

Begin by listing all available bounties.`, filterNote)

	return &mcp.GetPromptResult{
		Description: "Hunt for bounties in the Gitopia ecosystem",
		Messages: []*mcp.PromptMessage{
			{Role: "user", Content: &mcp.TextContent{Text: prompt}},
		},
	}, nil
}
