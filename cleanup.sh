#!/bin/bash

# Define color codes
GREEN=$(tput setaf 2)
YELLOW=$(tput setaf 3)
RED=$(tput setaf 1)
BLUE=$(tput setaf 4)
NC=$(tput sgr0) # No Color

_cleanup() {
    # Handle -h option separately
    if [[ "$1" == "-h" ]]; then
        echo "Usage: swissgit cleanup [options]"
        echo "Options:"
        echo "  -h                      Show this help status_line and exit."
        echo "  -a                      Apply cleanup to all repositories in the target directory, including subdirectories."
        echo "  -d                      Drop all changes in the repositories. This will reset the repository to the last commit and remove untracked files."
        echo "  -f <target_dir>         Specify the target directory where the repositories are located. Defaults to the current directory."
        echo "  -s                      Rebase with autostash"
        return 0
    fi

    local drop_changes=0
    local stash_changes=0
    local target_dir="."
    local all_repos_flag=false

    while getopts ":dasf:" opt; do
        case ${opt} in
        a)
            all_repos_flag=true
            ;;
        d)
            drop_changes=1
            ;;

        f)
            target_dir="$OPTARG"
            ;;
        s)
            stash_changes=1
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit cleanup [-h][-d][-a][-s] [-f <target_dir>] "
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit cleanup [-h][-d][-a][-s] [-f <target_dir>] "
            return 1
            ;;
        esac
    done
    shift $((OPTIND - 1))

    if [[ $all_repos_flag == true ]]; then
        find_options=()
    else
        find_options=(-maxdepth 0)
    fi

    # Find directories and iterate through them
    while IFS= read -r -d '' dir; do
        if [[ -d "$dir/.git" ]]; then
            (
                cd "$dir"

                # Check if there are changes (including untracked files)
                changes=""
                if git status --porcelain | grep -Eq '^.M|^.D|^\?\?'; then
                    if [[ $drop_changes -eq 1 ]]; then
                        git reset --hard HEAD >/dev/null 2>&1 && git clean -xdf >/dev/null 2>&1
                        changes="${RED}Dropped all changes${NC}"
                    else
                        changes=$(git status --short | awk '{print $1}' | sort | uniq -c | awk '{print $2$1}' | tr -d '\n' |
                            sed -e "s/??/${GREEN}U${NC}/g" -e "s/D/${RED}D${NC}/g" -e "s/M/${YELLOW}M${NC}/g")
                    fi
                fi

                # Update the main branch, prune and count branches
                if [[ -z $changes || $changes == "${RED}Dropped all changes${NC}" ]]; then
                    git switch main >/dev/null 2>&1
                    git fetch --prune >/dev/null 2>&1
                    if [[ $stash_changes -eq 1 ]]; then
                        git pull --rebase --autostash >/dev/null 2>&1
                    else
                        git pull >/dev/null 2>&1
                    fi
                fi

                current_branch=$(git rev-parse --abbrev-ref HEAD)
                pruned_branches=$(git branch --merged main | grep -v '^\* main$' | wc -l | awk '{$1=$1};1' || echo -e "${RED}Error${NC}")
                if [[ $pruned_branches == "Error" ]]; then
                    pruned_branches="${RED}Error${NC}"
                else
                    git branch --merged main | grep -v '^\* main$' | xargs -n 1 git branch -d >/dev/null 2>&1
                fi
                branches=$(git branch | wc -l | awk '{$1=$1};1')

                # Initialize a variable to construct the entire status_line
                local status_line=""

                if [[ -z $changes && $branches -eq 1 && $current_branch == "main" && $pruned_branches == 0 ]]; then
                    status_line="${GREEN}$dir: ${NC}"
                else
                    status_line="${RED}$(basename "$PWD"): ${NC}"
                    if [[ -n $changes ]]; then
                        status_line+="${BLUE}Non Committed Changes:${NC} $changes "
                    fi
                    if [[ $branches -ne 1 ]]; then
                        status_line+="${YELLOW} Too many branches: $branches${NC} "
                    fi
                    if [[ $current_branch != "main" ]]; then
                        status_line+="${BLUE} Current Branch: $current_branch${NC} "
                    fi
                    if [[ $pruned_branches -ne 0 ]]; then
                        status_line+="${YELLOW} Pruned Branches: $pruned_branches${NC}"
                    fi
                fi

                # Print the entire status_line for this directory at once
                echo -e "$status_line"
            ) &
        fi
    done < <(find "$target_dir" "${find_options[@]}" -type d -print0)

    wait # Wait for all background processes to finish
}
