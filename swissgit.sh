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
  status                Recursively checks the status of repositories. If current dir is a git repo, it will only check that repo. Use 'swissgit status -h' for more information.
  branches              Recursively checks the branch status of repositories. If current dir is a git repo, it will only check that repo. Use 'swissgit branches -h' for more information.
  cleanup               Clean up your repositories. Check out and update main, drop merged branches and drop no longer needed changes. Use 'swissgit cleanup -h' for more information.
  clone                 Clone a teams repositories with SSH. Requires a personal access token. Use 'swissgit clone -h' for more information.
  commit                Create and push a commit on the current branch or a new one. Without a PR. Use 'swissgit commit -h' for more information.
  pullrequest           Create a pull request. Creates a branch, commits all your changes, and creates a pull request. Use 'swissgit pullrequest -h' for more information.
  help                  Show this help message and exit
EOF
}

# Dispatch commands based on user input
swissgit() {
    unsetopt MONITOR # or set +m
    if [ "$#" -eq 0 ]; then
        echo "Error: No command provided. Use 'swissgit help' for usage information."
        return 1
    fi

    local command="$1"
    shift

    case "$command" in
    status)
        _status "$@"
        ;;
    branches)
        _branches "$@"
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
