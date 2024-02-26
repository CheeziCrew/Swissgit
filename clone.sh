#!/bin/bash

_clone() {
    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit clone [options]"
        echo "Options:"
        echo "  -h                      Show this help message and exit."
        echo "  -o <organization>       Specify the GitHub organization."
        echo "  -t <team>               Specify the team within the organization."
        echo "  -k <token>              Specify the GitHub personal access token."
        echo "  -d <target_dir>         Specify the target directory for cloning. Defaults to the current directory."
        return 0
    fi

    organization=""
    team=""
    token=""
    target_dir="."

    while getopts ":o:t:k:d:" opt; do
        case ${opt} in
        o)
            organization="$OPTARG"
            ;;
        t)
            team="$OPTARG"
            ;;
        k)
            token="$OPTARG"
            ;;
        d)
            target_dir="$OPTARG"
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit clone -o <organization> -t <team> -k <token> [-d <target_dir>]"
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit clone -o <organization> -t <team> -k <token> [-d <target_dir>]"
            return 1
            ;;
        esac
    done
    shift $((OPTIND - 1))

    # Check if required options are provided
    if [[ -z $organization || -z $team || -z $token ]]; then
        echo "Error: Required options are missing. Usage: swissgit clone -o <organization> -t <team> -k <token> [-d <target_dir>]" >&2
        return 1
    fi

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
