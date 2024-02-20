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

usage() {
    cat <<-EOF
Usage: swissgit [-h | --help] [-s | --status] [-b | --branches] [-c | --clean [-a] [-b] [folder]]
                [-p | --pullrequest [-a] <branchname> <commit_message> [PR_body]]

Options:
  -h, --help             Show this help message and exit
  -s, --status           Checks recursively the status of all repositories
  -b, --branches         Checks recursively the branch status of all repositories
  -c, --clean [-a] [-d] [folder]
                         Clean untracked files. Use -a to clean all, -d to drop local changes, and [folder] to specify a folder.
  -p, --pullrequest [-a] <branchname> <commit_message> [PR_body]
                         Create a pull request. Use -a for recursively doing for all subdirectories.
                         Creates a branch, commits all your changes and creates a pull pullrequest.


EOF
}

# Dispatch commands based on user input
swissgit() {
    case "$1" in
    -s | --status)
        status
        ;;
    -b | --branches)
        branches
        ;;

    -c | --cleanup)
        cleanup "$2" "$3" "$4"
        ;;
    -p | --pullrequest)
        if [ "$#" -lt 3 ]; then
            echo "Usage: swissgit [-p | --pullrequest] [-a] <branchname> <commit_message> [PR_body]"
            return 1
        fi
        pullrequest "$2" "$3" "$4" "$5"
        ;;
    -h | --help)
        usage
        ;;
    *)
        usage
        ;;
    esac
}
