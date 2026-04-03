# Gitopia MCP Tools Reference (57 tools)

## Context Management (4)
| Tool | Description |
|------|-------------|
| `get_user_context` | Get identity, wallet address, and DAO memberships |
| `set_active_dao` | Switch operations to a DAO context |
| `refresh_user_context` | Reload context after account changes |
| `claim_fee_grant` | Claim fee grant from Gitopia faucet |

## User (1)
| Tool | Description |
|------|-------------|
| `create_user` | Create on-chain Gitopia user for wallet |

## Repository Management (7)
| Tool | Description |
|------|-------------|
| `list_repos` | List repositories for a user or DAO |
| `create_repo` | Create remote repository on Gitopia |
| `get_repo` | Get repository details (id, description, forks) |
| `list_branches` | List branches with SHA |
| `get_file_contents` | Read file from remote repo without cloning |
| `fork_repository` | Fork a repository |
| `toggle_repository_forking` | Enable/disable forking |

## Metadata (5)
| Tool | Description |
|------|-------------|
| `list_tags` | List version tags |
| `list_commits` | Browse commit history for a branch |
| `list_releases` | List published releases |
| `list_labels` | List repository labels |
| `create_release` | Publish a release for a tagged version |

## Labels (2)
| Tool | Description |
|------|-------------|
| `create_label` | Create label (on-chain) |
| `delete_label` | Delete label (on-chain, irreversible) |

## Issue Management (5)
| Tool | Description |
|------|-------------|
| `list_issues` | List issues for a repository |
| `get_issue` | Get full issue details (labels, assignees, bounties) |
| `create_issue` | Create issue (on-chain) |
| `comment_on_issue` | Comment on issue (on-chain) |
| `update_issue` | Update state, labels, assignees (atomic single tx) |

## Pull Request Management (6)
| Tool | Description |
|------|-------------|
| `list_pull_requests` | List PRs for a repository |
| `get_pull_request` | Get full PR details (head, base, reviewers) |
| `get_pull_request_diff` | Get unified diff with per-file stats |
| `create_pull_request` | Create PR (on-chain, supports issue linking) |
| `comment_on_pull_request` | Comment on PR (general or inline) |
| `merge_pull_request` | Merge PR (on-chain) |

## Git Operations (5)
| Tool | Description |
|------|-------------|
| `git_clone` | Clone Gitopia repo to local workspace |
| `git_push` | Push commits to remote |
| `create_feature_branch` | Create and checkout new branch |
| `sync_with_remote` | Fetch and merge/rebase from remote |
| `commit_and_push_changes` | Stage + commit + push in one step |

## Workflow Orchestrators (3)
| Tool | Description |
|------|-------------|
| `bootstrap_repo` | Create remote + init local + README + push |
| `create_feature_branch_pr` | Branch + changes + commit + push + PR |
| `update_feature_branch` | Add commits to existing branch/PR |

Note: These are in the `workflow` toolset (hidden when `TOOLSETS=core`).

## DAO (2)
| Tool | Description |
|------|-------------|
| `get_dao` | Get DAO details (group_id, group_policy_address) |
| `create_dao` | Create new DAO with members and voting |

## DAO Governance (7)
| Tool | Description |
|------|-------------|
| `dao_list_members` | List members and voting weights (accepts `dao` name) |
| `dao_list_proposals` | List proposals (accepts `dao` name) |
| `dao_get_proposal` | Get proposal details with tally results |
| `dao_update_members` | Add/remove members (accepts `dao` name) |
| `dao_submit_proposal` | Create governance proposal (accepts `dao` name) |
| `dao_vote` | Cast vote (yes/no/abstain/no_with_veto) |
| `dao_exec` | Execute a passed proposal |

## Bounties (6)
| Tool | Description |
|------|-------------|
| `list_bounties` | Discover available bounties |
| `get_bounty` | Get bounty details (amount, expiry, state) |
| `create_bounty` | Attach reward to an issue |
| `update_bounty` | Extend bounty expiry |
| `close_bounty` | Deactivate bounty |
| `delete_bounty` | Permanently remove bounty |

## Batch & Approval (4)
| Tool | Description |
|------|-------------|
| `batch_execute` | Execute up to 10 ops in single atomic tx |
| `confirm_transaction` | Broadcast pending transaction (approval mode) |
| `reject_transaction` | Cancel pending transaction |
| `list_pending_transactions` | View pending transactions |
