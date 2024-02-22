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
Usage: swissgit COMMAND [OPTIONS]

Commands:
  status                Recursively checks the status of all repositories
  branches              Recursively checks the branch status of all repositories
  cleanup [-a] [-d] [folder]
                        Clean untracked files. Use -a to clean all, -d to drop local changes, and [folder] to specify a folder.
  clone <org> <team> <github_token> [target_dir]
                        Clone a team's repositories with SSH. Requires a personal access token.
  commit <commit_message> [branchname] 
                        Create and push a commit on the current branch or a new one. Without a PR       
  pullrequest [-a] <branchname> <commit_message> [PR_body]
                        Create a pull request. Use -a for recursively doing for all subdirectories. Creates a branch, commits all your changes, and creates a pull request.
  help                  Show this help message and exit
EOF
}

# Dispatch commands based on user input
swissgit() {
    if [ "$#" -eq 0 ]; then
        echo "Error: No command provided. Use 'swissgit help' for usage information."
        return 1
    fi

    local command="$1"
    shift

    case "$command" in
    status)
        _status
        ;;
    branches)
        _branches
        ;;
    cleanup)
        _cleanup "$@"
        ;;
    clone)
        _clone "$@"
        ;;
    commit)
        _commit "$@"
        ;;
    pullrequest)
        _pullrequest "$@"
        ;;
    help)
        _usage
        ;;
    *)
        echo "Invalid command. Use 'swissgit help' for usage information."
        ;;
    esac
}
