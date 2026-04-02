# Gitopia Bounty Workflow

Guided workflow to discover and work on bounties.

## Steps

1. **List available bounties** using `gitopia_list_bounties` to find open bounties
2. **Get bounty details** using `gitopia_get_bounty` to evaluate reward and requirements
3. **Read the linked issue** using `gitopia_get_issue` to understand what needs to be done
4. **Execute the fix** — follow the contribute workflow (clone, branch, fix, test, PR)
5. **Reference the bounty** in your PR description

## Example

```
1. gitopia_list_bounties limit=10        -> list of bounties with amounts
2. gitopia_get_bounty 42                 -> full details, expiry, linked issue
3. gitopia_get_issue owner repo 7        -> issue description and context
4. Follow /gitopia-contribute workflow to submit the fix
5. Reference bounty #42 in the PR description
```

## Tips

- Filter bounties by checking `state` and `expire_at` fields
- Higher-value bounties often have more complex requirements
- Check if someone else is already working on the bounty (look at issue comments)
- Make sure your fix fully addresses the issue requirements before submitting
