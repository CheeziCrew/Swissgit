#!/bin/bash

_clone() {

    organization=$1
    team=$2
    token=$3
    target_dir=$4

    # Set default clone directory if not provided
    target_dir="${target_dir:-SWISSGIT_DEFAULT_CLONE}"
    # Make clone dir if not exists
    mkdir -p "$target_dir" && cd "$target_dir"

    # GitHub API endpoint to list repositories for a team
    url="orgs/${organization}/teams/${team}/repos"

    # Send a GET request to the GitHub API using `gh` CLI
    response=$(gh api --paginate "$url" -H "Authorization: token ${token}" -H "Accept: application/vnd.github.v3+json")
    # Extract clone URLs from the response and clone repositories
    echo "$response" | jq -r '.[].ssh_url' | while read -r ssh_url; do
        echo "Cloning repository: $ssh_url"
        git clone "$ssh_url" >/dev/null 2>&1
    done

    echo "Cloning completed successfully."
}
