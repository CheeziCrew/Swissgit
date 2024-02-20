#!/bin/bash

pullrequest() {
    local all_repos_flag=false

    # Check for flags
    while [[ "$#" -gt 0 ]]; do
        case "$1" in
        -a)
            all_repos_flag=true
            shift
            ;;
        *)
            # Skip this case if it starts with a dash
            if [[ "$1" == -* ]]; then
                echo "Unknown flag: $1"
                return 1
            else
                # If it's not a flag, it's a parameter
                break
            fi
            ;;
        esac
    done

    # Check if branchname and commit message are provided
    if [ $# -lt 2 ]; then
        echo "Usage: pullrequest [-a] <branchname> <commit_message> [PR_body]"
        return 1
    fi

    local branchname="$1"
    local commit_message="$2"
    local pr_body="$3"

    if [[ $all_repos_flag == true ]]; then
        for dir in */; do
            if [[ -d "$dir/.git" ]]; then
                dir="${dir%/}" # Remove trailing slash
                (
                    cd "$dir" || return
                    pullrequest "$branchname" "$commit_message" "$pr_body"
                )
            fi
        done
        wait # Wait for all background processes to finish
    else
        # Check if there are changes to commit
        if [[ -z $(git status --porcelain) ]]; then
            echo "No changes to commit. Aborting."
            return 1
        fi

        # Check if the branch already exists
        if git rev-parse --verify "$branchname" >/dev/null 2>&1; then
            echo "Branch '$branchname' already exists. Aborting."
            return 1
        fi

        # Checkout a new branch
        git checkout -b "$branchname" >/dev/null 2>&1

        # Check if the branch creation was successful
        if [ $? -ne 0 ]; then
            echo "Failed to create branch '$branchname'. Aborting."
            return 1
        fi

        # Add all changes
        git add . >/dev/null 2>&1

        # Commit changes
        git commit -m "$commit_message" >/dev/null 2>&1

        # Push changes
        git push origin "$branchname" >/dev/null 2>&1

        # Check if the push was successful
        if [ $? -ne 0 ]; then
            echo "Failed to push changes to the remote repository. Trying to pull latest changes..."
            git pull origin "$branchname" # Attempt to pull latest changes from remote
            if [ $? -eq 0 ]; then
                echo "Successfully pulled latest changes. Pushing again"
                git push origin "$branchname"
            else
                echo "Failed to pull latest changes. Please resolve conflicts manually."
            fi
            return 1
        fi

        # Create PR using GitHub CLI and capture the output
        pr_output=$(gh pr create --title "$commit_message" --body "$pr_body" --base main --head "$branchname" 2>&1)

        # Extract the URL from the output
        pr_url=$(echo "$pr_output" | grep -o 'https://github.com/[^\"]*')

        # Display the URL
        echo "Pull request created: $pr_url"
    fi
}
