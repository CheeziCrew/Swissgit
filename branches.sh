#!/bin/bash

# Define terminal color codes
GREEN=$(tput setaf 2)
BLUE=$(tput setaf 4)
RED=$(tput setaf 1)
YELLOW=$(tput setaf 3)
NC=$(tput sgr0) # No Color

_branches() {
    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit branches [options]"
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

                local repo_name=$(basename "$PWD")
                local local_branches=$(cd "$dir" && git branch --list | wc -l | tr -d ' ')
                local remote_branches=$(cd "$dir" && git branch -r | grep -v '\->' | wc -l | tr -d ' ')
                local main_branch=$(cd "$dir" && git symbolic-ref --short -q HEAD)
                local all_branches=($(cd "$dir" && git branch -r | grep -v '\->' | grep -v 'HEAD' | sed -e 's/^[ \t]*//' | awk '{print $1}'))
                local stale_branches=()
                local other_branches=()

                # Count stale branches
                for branch in "${all_branches[@]}"; do
                    last_commit_timestamp=$(cd "$dir" && git log -1 --format="%at" "$branch")
                    current_timestamp=$(date +%s)
                    age=$((current_timestamp - last_commit_timestamp))
                    if [[ "$age" -gt "10368000" && "${branch#origin/}" != "$main_branch" ]]; then # 120 days in seconds
                        stale_branches+=("${branch#origin/}")
                    elif [[ "${branch#origin/}" != "$main_branch" ]]; then
                        other_branches+=("${branch#origin/}")
                    fi
                done

                local status_line="${repo_name}: "

                # Check if there are no other branches or stale branches
                if [[ ${#other_branches[@]} -eq 0 && ${#stale_branches[@]} -eq 0 ]]; then
                    status_line+="${GREEN}󰞑${NC}"
                else
                    status_line+="(${BLUE}L${local_branches}/R${remote_branches}${NC}): "

                    # Print main branch
                    if [[ "${main_branch#origin/}" == "main" ]]; then
                        status_line+="${BLUE}${main_branch#origin/}${NC}; "
                    else
                        status_line+="${RED}${main_branch#origin/}${NC}; "
                    fi

                    # Print other branches
                    if [[ ${#other_branches[@]} -gt 0 ]]; then
                        status_line+="${YELLOW}"
                        local count=0
                        for branch in "${other_branches[@]}"; do
                            if [[ $count -ne 0 ]]; then
                                status_line+=", "
                            fi
                            status_line+="${branch}"
                            count=$((count + 1))
                            # Print only five branches, followed by the number of remaining branches
                            if [[ $count -eq 5 && ${#other_branches[@]} -gt 5 ]]; then
                                status_line+=" and $((${#other_branches[@]} - 5)) more"
                                break
                            fi
                        done
                        status_line+="${NC}; "
                    fi

                    # Print stale branches
                    if [[ ${#stale_branches[@]} -gt 0 ]]; then
                        status_line+="${RED} ${#stale_branches[@]} stale branches${NC}"
                    fi
                fi

                # Echo the fully assembled status line
                echo "$status_line"
            else
                echo "Warning: $dir is not a valid Git repository. Skipping..."
            fi
        } &
    done
    wait # Wait for all background jobs to finish
}
