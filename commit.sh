#!/bin/bash

_commit() {

    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit commit [options]"
        echo "Options:"
        echo "  -h                      Show this help message and exit."
        echo "  -a                      Apply the operation to all repositories in the current directory."
        echo "  -c <commit_message>     Specify the commit message."
        echo "  -b <branchname>         Specify the branch name to commit to. If the branch does not exist, it will be created."
        echo "  -n                      Do not push the commit to the remote repository."
        return 0
    fi

    # Initialize variables
    local push=true
    local all_repos_flag=false

    # Get options
    while getopts ":ac:b:n" opt; do
        case ${opt} in
        a)
            all_repos_flag=true
            ;;
        c)
            commit_message="$OPTARG"
            ;;
        b)
            branchname="$OPTARG"
            ;;
        n)
            push=false
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit commit [-h] [-a] -c <commit_message> [-b <branchname>]"
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit commit [-h] [-a] -c <commit_message> [-b <branchname>]"
            return 1
            ;;
        esac
    done
    shift $((OPTIND - 1))

    if [[ $all_repos_flag == true ]]; then
        for dir in */; do
            if [[ -d "$dir/.git" ]]; then
                dir="${dir%/}" # Remove trailing slash
                (
                    cd "$dir" || return
                    _commit "$branchname" "$commit_message"
                )
            fi
        done
    else
        # Check if both commit_message provided
        if [[ -z $commit_message ]]; then
            echo "Error: Commit message is required." >&2
            echo "Usage: swissgit commit [-h] [-a] -c <commit_message> [-b <branchname>]"
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

        if [[ -z $branchname ]]; then
            branchname=$(git branch --show-current 2>/dev/null)
        fi

        # Add all changes
        git add . >/dev/null 2>&1

        # Commit changes
        git commit -m "$branchname: $commit_message" >/dev/null 2>&1

        if [[ $push == true ]]; then
            # Push changes
            git push >/dev/null 2>&1

            # Check if the push was successful
            if [ $? -ne 0 ]; then
                echo "Failed to push changes to the remote repository. Trying to pull latest changes..."
                git pull && git push || {
                    echo "Failed to push changes. Please check your connection or permissions, and try again."
                    return 1
                }
            fi
        fi
    fi

}
