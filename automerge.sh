#!/bin/bash

# Use the current directory as the base directory
base_dir=$(pwd)

# Recursively find the folder with a PR that starts with UF-10098
for repo in $(find "$base_dir" -type d -name ".git"); do
    repo_dir=$(dirname "$repo")
    cd "$repo_dir" || continue

    # Fetch open PRs starting with UF-10098
    pr_number=$(gh pr list --limit 100 --search "UF-10098" --json number --jq '.[0].number')

    if [ -n "$pr_number" ]; then
        echo "Found PR #$pr_number in $repo_dir, enabling auto-merge with merge commit..."
        gh pr merge "$pr_number" --auto --merge --delete-branch=true
    else
        echo "No matching PR in $repo_dir"
    fi
done
