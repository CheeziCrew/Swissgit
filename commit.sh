#!/bin/bash

_commit() {

    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit [-h] -c <commit_message> -b <branchname>"
        return 0
    fi

    # Initialize variables
    local commit_message=""
    local branchname=""

    # Get options
    while getopts ":c:b:" opt; do
        case ${opt} in
        c)
            commit_message="$OPTARG"
            ;;
        b)
            branchname="$OPTARG"
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit [-h] -c <commit_message> -b <branchname>"
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit [-h] -c <commit_message> -b <branchname>"
            return 1
            ;;
        esac
    done
    shift $((OPTIND - 1))

    # Check if both commit_message and branchname are provided
    if [[ -z $commit_message || -z $branchname ]]; then
        echo "Error: Both commit message and branch name are required." >&2
        echo "Usage: swissgit [-h] -c <commit_message> -b <branchname>"
        return 1
    fi

    if [[ -z $(git status --porcelain) ]]; then
        echo "No changes to commit. Aborting."
        return 1
    fi

    if [[ -n $branchname ]]; then
        # Check if the branch already exists
        if git rev-parse --verify "$branchname" >/dev/null 2>&1; then
            echo "Branch '$branchname' already exists. Checking out existing branch."
            git checkout "$branchname" >/dev/null 2>&1 || {
                echo "Failed to checkout branch '$branchname'. Aborting."
                return 1
            }
        else
            # Checkout a new branch
            git checkout -b "$branchname" >/dev/null 2>&1 || {
                echo "Failed to create branch '$branchname'. Aborting."
                return 1
            }
        fi
    fi

    # Add all changes
    git add . >/dev/null 2>&1

    # Commit changes
    git commit -m "$commit_message" >/dev/null 2>&1

    # Push changes
    git push >/dev/null 2>&1

    # Check if the push was successful
    if [ $? -ne 0 ]; then
        echo "Failed to push changes to the remote repository. Trying to pull latest changes..."
        git pull && git push || {
            echo "Failed to pull latest changes. Please resolve conflicts manually."
            return 1
        }
    fi

}
