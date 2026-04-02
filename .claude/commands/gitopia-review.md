# Gitopia PR Review Workflow

Guided workflow to review a pull request on Gitopia.

## Steps

1. **Get PR details** using `gitopia_get_pull_request` to understand the change
2. **Clone the repository** using `git_clone` if not already local
3. **Read changed files** using `read_file` to inspect the code
4. **Search for patterns** using `search_code` to check for regressions
5. **Leave review comments** using `gitopia_comment_on_pull_request`

## Example

```
1. gitopia_get_pull_request owner repo 5  -> PR title, description, head/base branches
2. git_clone owner/repo -> local-repo
3. read_file local-repo/src/changed_file.go
4. search_code local-repo "functionName"
5. gitopia_comment_on_pull_request owner repo 5 "LGTM - code looks clean"
```

## Review Checklist

- Does the code compile and pass tests?
- Are there any security concerns?
- Is the change well-documented?
- Does it follow existing conventions?
- Are edge cases handled?
