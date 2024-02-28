#!/bin/bash

_status() {
    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit status [options]"
        echo "Options:"
        echo "  -h                      Show this help message and exit."
        return 0
    fi

    # Check if the current directory is a git repository
    if [ -d ".git" ]; then
        dirs=(".")
    else
        local dirs=($(find . -maxdepth 1 -mindepth 1 -type d))
    fi
    # Define terminal color codes
    GREEN=$(tput setaf 2)
    YELLOW=$(tput setaf 3)
    RED=$(tput setaf 1)
    BLUE=$(tput setaf 4)
    NC=$(tput sgr0) # No Color

    for dir in "${dirs[@]}"; do
        if [[ -d "$dir/.git" ]]; then
            local repo_name=$(basename "$dir")
            local status_output=$(cd "$dir" && git status --porcelain)
            local branch=$(cd "$dir" && git symbolic-ref --short -q HEAD)
            local ahead_behind=$(cd "$dir" && git rev-list --left-right --count HEAD...@{u} 2>/dev/null)
            local ahead=$(echo "$ahead_behind" | awk '{print $1}')
            local behind=$(echo "$ahead_behind" | awk '{print $2}')
            local changed=$(echo "$status_output" | grep -c '^ M\|^M ')
            local new=$(echo "$status_output" | grep -c '^ A\|^A ')
            local deleted=$(echo "$status_output" | grep -c '^ D\|^D ')
            local untracked=$(echo "$status_output" | grep -c '^??')

            echo -n "${repo_name}: "

            # Print branch name if available
            if [[ -n $branch ]]; then
                # Print main branch
                if [[ "$branch" == "main" ]]; then
                    echo -n "${GREEN}$branch ${NC}"
                else
                    echo -n "${YELLOW}$branch ${NC}"
                fi

            fi

            # Print commits ahead/behind if available
            if [[ -n $ahead && $ahead -gt 0 ]] || [[ -n $behind && $behind -gt 0 ]]; then
                echo -n "(${GREEN}${ahead}/${RED}${behind}) ${NC}"
            fi

            # Print changes with colors
            if [[ $changed -gt 0 ]]; then
                echo -n "${YELLOW}(M) ${changed}  ${NC}"
            fi
            if [[ $new -gt 0 ]]; then
                echo -n "${GREEN}(A) ${new}  ${NC}"
            fi
            if [[ $deleted -gt 0 ]]; then
                echo -n "${RED}(D) ${deleted}  ${NC}"
            fi
            if [[ $untracked -gt 0 ]]; then
                echo -n "${BLUE}(U) ${untracked}  ${NC}"
            fi

            # Add a green 󰞑 if no changes and not ahead/behind
            if [[ $changed -eq 0 && $new -eq 0 && $deleted -eq 0 && $untracked -eq 0 && -z $ahead && -z $behind ]]; then
                echo -n "${GREEN}󰞑${NC}"
            fi

            echo ""
        fi
    done
}
