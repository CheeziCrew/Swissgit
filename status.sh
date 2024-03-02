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
        {
            # Check if the directory exists and is a valid Git repository
            if [[ -d "$dir/.git" ]] && (cd "$dir" && git rev-parse --git-dir >/dev/null 2>&1); then
                # Attempt to change directory and fetch all branches
                if ! cd "$dir" >/dev/null 2>&1; then
                    echo "Error: Failed to change directory to $dir. Skipping..."
                    continue
                fi

                if ! git fetch --all >/dev/null 2>&1; then
                    echo "Error: Failed to fetch Git repositories in $dir. Skipping..."
                    cd - >/dev/null 2>&1 # Return to the previous directory
                    continue
                fi

                cd - >/dev/null 2>&1 # Return to the previous directory

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

                local status_line="$(cd "$dir" && basename "$PWD"): "

                # Build the status line
                if [[ -n $branch ]]; then
                    if [[ "$branch" == "main" ]]; then
                        status_line+="${GREEN}$branch ${NC}"
                    else
                        status_line+="${YELLOW}$branch ${NC}"
                    fi
                fi

                if [[ -n $ahead && $ahead -gt 0 ]] || [[ -n $behind && $behind -gt 0 ]]; then
                    status_line+="(${GREEN}${ahead}/${RED}${behind}) ${NC}"
                fi

                if [[ $changed -gt 0 ]]; then
                    status_line+="${YELLOW}(M) ${changed}  ${NC}"
                fi
                if [[ $new -gt 0 ]]; then
                    status_line+="${GREEN}(A) ${new}  ${NC}"
                fi
                if [[ $deleted -gt 0 ]]; then
                    status_line+="${RED}(D) ${deleted}  ${NC}"
                fi
                if [[ $untracked -gt 0 ]]; then
                    status_line+="${BLUE}(U) ${untracked}  ${NC}"
                fi

                if [[ $changed -eq 0 && $new -eq 0 && $deleted -eq 0 && $untracked -eq 0 && -z $ahead && -z $behind ]]; then
                    status_line+="${GREEN}󰞑${NC}"
                fi

                # Echo the fully assembled status line
                echo "$status_line"
            fi
        } &
    done
    wait # Wait for all background jobs to finish
}
