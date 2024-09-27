#!/bin/bash

create_pull_request_in_repo() {
    local branchname="$1"
    local commit_message="$2"
    local pr_body="$3"

    # Check if there are changes to commit
    if [[ -z $(git status --porcelain) ]]; then
        echo "$(basename "$PWD"): No changes to commit. Aborting."
        return 1
    fi

    # Check if the branch already exists or create a new branch
    git rev-parse --verify "$branchname" >/dev/null 2>&1 && {
        echo "$(basename "$PWD"): Branch '$branchname' already exists. Aborting."
        return 1
    } || git checkout -b "$branchname" >/dev/null 2>&1 || {
        echo "$(basename "$PWD"): Failed to create branch '$branchname'. Aborting."
        return 1
    }

    # Add all changes
    git add . >/dev/null 2>&1 || {
        echo "$(basename "$PWD"): Failed to add changes. Aborting."
        return 1
    }

    # Commit changes
    git commit -m "$branchname: $commit_message" >/dev/null 2>&1 || {
        echo "$(basename "$PWD"): Commit failed. Aborting."
        return 1
    }

    # Push changes
    git push origin "$branchname" >/dev/null 2>&1 || {
        echo "$(basename "$PWD"): Failed to push changes to the remote repository. Trying to pull latest changes..."
        git pull origin "$branchname" && git push origin "$branchname" || {
            echo "$(basename "$PWD"): Failed to pull latest changes. Please resolve conflicts manually."
            return 1
        }
    }

    # Create PR using GitHub CLI and capture the output
    pr_output=$(gh pr create --title "$branchname: $commit_message" --body "$pr_body" --base main --head "$branchname" 2>&1)
    if [ $? -ne 0 ]; then
        echo "$(basename "$PWD"): Failed to create pull request. Error: $pr_output"
        return 1
    fi

    # Extract the URL from the output
    pr_url=$(echo "$pr_output" | grep -o 'https://github.com/[^"]*')

    # Display the URL
    echo "Pull request created: $pr_url"
}

_pullrequest() {
    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit pullrequest [options]"
        echo "Options:"
        echo "  -h                      Show this help message and exit."
        echo "  -a                      Apply the operation to all repositories in the current directory."
        return 0
    fi

    local all_repos_flag=false
    local branchname=""
    local commit_message=""
    local pr_body=""

    while getopts ":ab:c:" opt; do
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
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit pullrequest [-a] -b <branchname> -c <commit_message>"
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit pullrequest [-a] -b <branchname> -c <commit_message>"
            return 1
            ;;
        esac
    done

    # Prompt for branch name and commit message if not provided
    if [[ -z $branchname ]]; then
        echo "Enter the branch name:"
        read branchname
    fi
    if [[ -z $commit_message ]]; then
        echo "Enter the commit message:"
        read commit_message
    fi

    # Read and process the PULL_REQUEST_TEMPLATE.md if not already provided
    if [[ -z $pr_body && -f "PULL_REQUEST_TEMPLATE.md" ]]; then
        pr_body=$(<PULL_REQUEST_TEMPLATE.md)

        # Prompt for types of changes
        echo "Select the types of changes (enter the numbers separated by spaces):"
        echo "1. Bug fix"
        echo "2. New feature"
        echo "3. Removed feature"
        echo "4. Code style update (formatting etc.)"
        echo "5. Refactoring (no functional changes, no api changes)"
        echo "6. Build related changes"
        echo "7. Documentation content changes"
        echo "Your choices:"
        read types_of_changes

        # Replace placeholders in the template using sed with escaping
        for choice in $types_of_changes; do
            case $choice in
            1) pr_body=$(echo "$pr_body" | sed 's/- \[ \] Bug fix/- [x] Bug fix/') ;;
            2) pr_body=$(echo "$pr_body" | sed 's/- \[ \] New feature/- [x] New feature/') ;;
            3) pr_body=$(echo "$pr_body" | sed 's/- \[ \] Removed feature/- [x] Removed feature/') ;;
            4) pr_body=$(echo "$pr_body" | sed 's/- \[ \] Code style update (formatting etc.)/- [x] Code style update (formatting etc.)/') ;;
            5) pr_body=$(echo "$pr_body" | sed 's/- \[ \] Refactoring (no functional changes, no api changes)/- [x] Refactoring (no functional changes, no api changes)/') ;;
            6) pr_body=$(echo "$pr_body" | sed 's/- \[ \] Build related changes/- [x] Build related changes/') ;;
            7) pr_body=$(echo "$pr_body" | sed 's/- \[ \] Documentation content changes/- [x] Documentation content changes/') ;;
            esac
        done

        # Prompt for breaking change
        echo "Does this PR introduce a breaking change? (y/n):"
        read breaking_change
        if [[ $breaking_change == "y" ]]; then
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] Yes (I have stepped the version number accordingly)/- [x] Yes (I have stepped the version number accordingly)/')
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] No/- [ ] No/')
        else
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] Yes (I have stepped the version number accordingly)/- [ ] Yes (I have stepped the version number accordingly)/')
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] No/- [x] No/')
        fi

        # Prompt for checklist items
        echo "Does your code follow the code style of this project? (y/n):"
        read code_style
        if [[ $code_style == "y" ]]; then
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] My code follows the code style of this project\./- [x] My code follows the code style of this project./')
        else
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] My code follows the code style of this project\./- [ ] My code follows the code style of this project./')
        fi

        echo "Have you updated the documentation accordingly (if applicable)? (y/n):"
        read documentation
        if [[ $documentation == "y" ]]; then
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have updated the documentation accordingly (if applicable)\./- [x] I have updated the documentation accordingly (if applicable)./')
        else
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have updated the documentation accordingly (if applicable)\./- [ ] I have updated the documentation accordingly (if applicable)./')
        fi

        echo "Have you added/updated tests to cover your changes (if applicable)? (y/n):"
        read tests_updated
        if [[ $tests_updated == "y" ]]; then
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have added\/updated tests to cover my changes (if applicable)\./- [x] I have added\/updated tests to cover my changes (if applicable)./')
        else
            pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have added\/updated tests to cover my changes (if applicable)\./- [ ] I have added\/updated tests to cover my changes (if applicable)./')
        fi
    fi

    shift $((OPTIND - 1))

    # Main logic for creating PR
    if [[ $all_repos_flag == true ]]; then
        for dir in */; do
            if [[ -d "$dir/.git" ]]; then
                dir="${dir%/}" # Remove trailing slash
                (
                    cd "$dir" || return
                    create_pull_request_in_repo "$branchname" "$commit_message" "$pr_body" &
                )
                sleep 5 # Sleep for 5 seconds
            fi
        done
    else
        create_pull_request_in_repo "$branchname" "$commit_message" "$pr_body"
    fi
}

_pullrequest "$@"
