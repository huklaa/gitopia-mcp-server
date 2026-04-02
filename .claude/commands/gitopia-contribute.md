# Gitopia Contribute Workflow

Guided workflow to contribute code to a Gitopia repository.

## Steps

1. **Clone the repository** using `git_clone` with the target repo URL
2. **Create a feature branch** using `create_feature_branch` with a descriptive branch name
3. **Make changes** — read files with `read_file`, write fixes with `write_file`
4. **Run tests** locally to verify the fix
5. **Commit and push** using `commit_and_push_changes` with a clear commit message
6. **Open a pull request** using `create_pull_request` linking to the relevant issue

## Example

```
1. git_clone owner/repo -> local-repo
2. create_feature_branch local-repo fix/issue-42
3. read_file + write_file to apply changes
4. commit_and_push_changes local-repo "fix: resolve issue #42"
5. create_pull_request owner repo "Fix issue #42" "Resolves #42" fix/issue-42 main
```

## Tips

- Use `list_issues` to find issues to work on
- Use `gitopia_get_issue` to read full issue context before starting
- Use `search_code` to find relevant code in the cloned repo
- Link issues in your PR with `issue_iids` parameter
