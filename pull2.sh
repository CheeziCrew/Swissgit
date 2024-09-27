#!/bin/bash

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

    while getopts ":a" opt; do
        case ${opt} in
        a)
            all_repos_flag=true
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit pullrequest [-a]"
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit pullrequest [-a]"
            return 1
            ;;
        esac
    done

    # Prompt for required inputs
    read -p "Enter the branch name: " branchname
    read -p "Enter the commit message: " commit_message

    # Read PULL_REQUEST_TEMPLATE.md
    if [[ -f "PULL_REQUEST_TEMPLATE.md" ]]; then
        pr_body=$(<PULL_REQUEST_TEMPLATE.md)
    else
        echo "Error: PULL_REQUEST_TEMPLATE.md not found." >&2
        return 1
    fi

    # Prompt for types of changes
    echo "Select the types of changes (enter the numbers separated by spaces):"
    echo "1. Bug fix"
    echo "2. New feature"
    echo "3. Removed feature"
    echo "4. Code style update (formatting etc.)"
    echo "5. Refactoring (no functional changes, no api changes)"
    echo "6. Build related changes"
    echo "7. Documentation content changes"
    read -p "Your choices: " types_of_changes

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
    read -p "Does this PR introduce a breaking change? (y/n): " breaking_change
    if [[ $breaking_change == "y" ]]; then
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] Yes (I have stepped the version number accordingly)/- [x] Yes (I have stepped the version number accordingly)/')
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] No/- [ ] No/')
    else
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] Yes (I have stepped the version number accordingly)/- [ ] Yes (I have stepped the version number accordingly)/')
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] No/- [x] No/')
    fi

    # Prompt for checklist items
    read -p "Does your code follow the code style of this project? (y/n): " code_style
    if [[ $code_style == "y" ]]; then
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] My code follows the code style of this project\./- [x] My code follows the code style of this project./')
    else
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] My code follows the code style of this project\./- [ ] My code follows the code style of this project./')
    fi

    read -p "Have you updated the documentation accordingly (if applicable)? (y/n): " documentation
    if [[ $documentation == "y" ]]; then
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have updated the documentation accordingly (if applicable)\./- [x] I have updated the documentation accordingly (if applicable)./')
    else
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have updated the documentation accordingly (if applicable)\./- [ ] I have updated the documentation accordingly (if applicable)./')
    fi

    read -p "Have you added/updated tests to cover your changes (if applicable)? (y/n): " tests_updated
    if [[ $tests_updated == "y" ]]; then
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have added\/updated tests to cover my changes (if applicable)\./- [x] I have added\/updated tests to cover my changes (if applicable)./')
    else
        pr_body=$(echo "$pr_body" | sed 's/- \[ \] I have added\/updated tests to cover my changes (if applicable)\./- [ ] I have added\/updated tests to cover my changes (if applicable)./')
    fi

    shift $((OPTIND - 1))

    # Your code to create the pull request goes here
    echo "Creating pull request with the following details:"
    echo "Branch: $branchname"
    echo "Commit Message: $commit_message"
    echo -e "PR Body: $pr_body"
}

_pullrequest "$@"
