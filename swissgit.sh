#!/bin/bash

# Check if the script is being executed directly
if [[ "${0}" = "${BASH_SOURCE[0]}" ]]; then
    echo "This script should not be executed directly. It should be sourced from your shell configuration file (e.g., .zshrc)."
    exit 1
fi

# Get the directory of the script
SCRIPT_DIR=$(dirname "$0")
# Source the Zsh function files
source "$SCRIPT_DIR/status.sh"
source "$SCRIPT_DIR/branches.sh"
source "$SCRIPT_DIR/cleanup.sh"
source "$SCRIPT_DIR/pullrequest.sh"
source "$SCRIPT_DIR/clone.sh"
source "$SCRIPT_DIR/commit.sh"

_usage() {
    cat <<-EOF
Usage: swissgit [-h | --help] [-s | --status] [-b | --branches] [-c | --clean [-a] [-d] [folder]]
                [-l | --clone <org> <team> <github_token> [target_dir]]
                [-o | --commit <commit_message> [branchname]]
                [-p | --pullrequest [-a] <branchname> <commit_message> [PR_body]]

Options:
  -h, --help             Show this help message and exit
  -s, --status           Checks recursively the status of all repositories
  -b, --branches         Checks recursively the branch status of all repositories
  -c, --clean [-a] [-d] [folder]
                         Clean untracked files. Use -a to clean all, -d to drop local changes, and [folder] to specify a folder.
  -l, --clone <org> <team> <github_token> [target_dir]
                         Clone a team's repositories with SSH.
                         Requires a personal access token.
  -o, --commit <commit_message> [branchname] 
                         Create and push a commit on the current branch or a new one. Without a PR       
  -p, --pullrequest [-a] <branchname> <commit_message> [PR_body]
                         Create a pull request. Use -a for recursively doing for all subdirectories.
                         Creates a branch, commits all your changes, and creates a pull request.
EOF
}

# Dispatch commands based on user input
swissgit() {
    case "$1" in
    -s | --status)
        _status
        ;;
    -b | --branches)
        _branches
        ;;

    -c | --cleanup)
        _cleanup "$2" "$3" "$4"
        ;;
    -o | --commit)
        if [ "$#" -lt 2 ]; then
            echo "Error: Missing parameters for clone"
            echo "Usage: swissgit [-c | --commit] <commit_message> [branchname]"
            return 1
        fi
        _commit "$2" "$3" "$4"
        ;;
    -l | --clone)
        if [ "$#" -lt 4 ]; then
            echo "Error: Missing parameters for clone"
            echo "Usage: swissgit [-l | --clone] <org> <team> <github_token> [target_dir]"
            return 1
        fi
        _clone "$2" "$3" "$4" "$5"
        ;;
    -p | --pullrequest)
        if [ "$#" -lt 3 ]; then
            echo "Error: Missing parameters for pull request"
            echo "Usage: swissgit [-p | --pullrequest] [-a] <branchname> <commit_message> [PR_body]"
            return 1
        fi
        _pullrequest "$2" "$3" "$4" "$5"
        ;;
    -h | --help)
        _usage
        ;;
    *)
        _usage
        ;;
    esac
}
