#!/bin/bash

# Define terminal color codes
GREEN=$(tput setaf 2)
BLUE=$(tput setaf 4)
RED=$(tput setaf 1)
YELLOW=$(tput setaf 3)
NC=$(tput sgr0) # No Color

branches() {
    local dirs=($(find . -maxdepth 1 -mindepth 1 -type d))

    for dir in "${dirs[@]}"; do
        if [[ -d "$dir/.git" ]]; then
            local repo_name=$(basename "$dir")
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

            # Check if there are no other branches or stale branches
            if [[ ${#other_branches[@]} -eq 0 && ${#stale_branches[@]} -eq 0 ]]; then
                echo -n "${repo_name}: ${GREEN}󰞑${NC}"
            else
                echo -n "${repo_name}: (${BLUE}L${local_branches}/R${remote_branches}${NC}): "

                # Print main branch
                if [[ "${main_branch#origin/}" == "main" ]]; then
                    echo -n "${BLUE}${main_branch#origin/}${NC}; "
                else
                    echo -n "${RED}${main_branch#origin/}${NC}; "
                fi

                # Print other branches
                if [[ ${#other_branches[@]} -gt 0 ]]; then
                    echo -n "${YELLOW}"
                    local count=0
                    for branch in "${other_branches[@]}"; do
                        if [[ $count -ne 0 ]]; then
                            echo -n ", "
                        fi
                        echo -n "${branch}"
                        count=$((count + 1))
                        # Print only five branches, followed by the number of remaining branches
                        if [[ $count -eq 5 && ${#other_branches[@]} -gt 5 ]]; then
                            echo -n " and $((${#other_branches[@]} - 5)) more"
                            break
                        fi
                    done
                    echo -n "${NC}; "
                fi

                # Print stale branches
                if [[ ${#stale_branches[@]} -gt 0 ]]; then
                    echo -n "${RED} ${#stale_branches[@]} stale branches${NC}"
                fi
            fi

            echo ""
        fi
    done
}
