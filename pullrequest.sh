#!/bin/bash

_pullrequest() {
    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit pullrequest [options]"
        echo "Options:"
        echo "  -h                      Show this help message and exit."
        echo "  -a                      Apply the operation to all repositories in the current directory."
        echo "  -b <branchname>         Specify the branch name for the pull request."
        echo "  -c <commit_message>     Specify the commit message."
        echo "  -p <pr_body>            Specify the body of the pull request."
        return 0
    fi

    local all_repos_flag=false
    local pr_body=""

    while getopts ":ab:c:p:" opt; do
        case ${opt} in
        a)
            all_repos_flag=true
            ;;
        b)
            branchname="$OPTARG"
            ;;
        c)
            commit_message="$OPTARG"
            ;;
        p)
            pr_body="$OPTARG"
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit pullrequest [-a] -b <branchname> -c <commit_message> [-p <pr_body>]"
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit pullrequest [-a] -b <branchname> -c <commit_message> [-p <pr_body>]"
            return 1
            ;;
        esac
    done

    # Check if branchname and commit_message are provided
    if [[ -z $branchname || -z $commit_message ]]; then
        echo "Error: Branchname and commit message are required." >&2
        echo "Usage: swissgit pullrequest [-a] -b <branchname> -c <commit_message> [-p <pr_body>]"
        return 1
    fi

    shift $((OPTIND - 1))

    if [[ $all_repos_flag == true ]]; then
        for dir in */; do
            if [[ -d "$dir/.git" ]]; then
                dir="${dir%/}" # Remove trailing slash
                (
                    cd "$dir" || return
                    _pullrequest "$branchname" "$commit_message" "$pr_body"
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

        # Check if the branch already exists or create a new branch
        git rev-parse --verify "$branchname" >/dev/null 2>&1 && {
            echo "Branch '$branchname' already exists. Aborting."
            return 1
        } || git checkout -b "$branchname" >/dev/null 2>&1 || {
            echo "Failed to create branch '$branchname'. Aborting."
            return 1
        }

        # Add all changes
        git add . >/dev/null 2>&1

        # Commit changes
        git commit -m "$commit_message" >/dev/null 2>&1

        # Push changes
        git push origin "$branchname" >/dev/null 2>&1

        # Check if the push was successful
        if [ $? -ne 0 ]; then
            echo "Failed to push changes to the remote repository. Trying to pull latest changes..."
            git pull origin "$branchname" && git push origin "$branchname" || {
                echo "Failed to pull latest changes. Please resolve conflicts manually."
                return 1
            }
        fi

        # Create PR using GitHub CLI and capture the output
        pr_output=$(gh pr create --title "$commit_message" --body "$pr_body" --base main --head "$branchname" 2>&1)

        # Extract the URL from the output
        pr_url=$(echo "$pr_output" | grep -o 'https://github.com/[^\"]*')

        # Display the URL
        echo "Pull request created: $pr_url"
    fi
}
