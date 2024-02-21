#!/bin/bash

_commit() {

    local commit_message="$1"
    local branchname="$2"

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
