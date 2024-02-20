#!/bin/bash

function clone() {
    # Check if the correct number of arguments are provided
    if [ "$#" -lt 3 ]; then
        echo "Usage: $0 <org> <team> <token> [folder]"
        exit 1
    fi

    # GitHub organization name
    organization=$1

    # GitHub team name
    team=$2

    # GitHub personal access token with repo permissions
    token=$3

    # Directory where you want to clone the repositories
    clone_directory=$4

    # Set default clone directory if not provided
    if [ -z "$clone_directory" ]; then
        clone_directory="new-scit"
    fi

    # GitHub API endpoint to list repositories for a team
    url="https://api.github.com/orgs/${organization}/teams/${team}/repos?page=3"

    # Send a GET request to the GitHub API using `gh`
    gh api "${url}" -H "Authorization: token ${token}" -H "Accept: application/vnd.github.v3+json" >response.json

    # Check if the response is successful
    if [ $? -eq 0 ]; then
        # Create the clone directory if it doesn't exist
        if [ ! -d "$clone_directory" ]; then
            mkdir -p "$clone_directory"
        fi

        # Change directory to the clone directory
        cd "$clone_directory" || exit

        # Clone each repository using SSH
        cat response.json | jq -r '.[].ssh_url' | while read -r ssh_url; do
            git clone "$ssh_url"
        done

        echo "Cloning completed successfully."
    else
        echo "Failed to fetch repositories. Status code: $?"
    fi

}
