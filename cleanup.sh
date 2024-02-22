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
        echo "Usage: swissgit cleanup [-d <drop_changes>] [-a <all_repos>] [-f <target_dir>]"
        return 0
    fi

    local drop_changes=0
    local target_dir="."
    local all_repos_flag=false

    while getopts ":daf:" opt; do
        case ${opt} in
        d)
            drop_changes=1
            ;;
        a)
            all_repos_flag=true
            ;;
        f)
            target_dir="$OPTARG"
            ;;
        \?)
            echo "Invalid option: -$OPTARG" >&2
            echo "Usage: swissgit clone -o <organization> -t <team> -k <token> [-d <target_dir>]"
            return 1
            ;;
        :)
            echo "Option -$OPTARG requires an argument." >&2
            echo "Usage: swissgit clone -o <organization> -t <team> -k <token> [-d <target_dir>]"
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
                    git pull >/dev/null 2>&1
                fi

                current_branch=$(git rev-parse --abbrev-ref HEAD)
                pruned_branches=$(git branch --merged main | grep -v '^\* main$' | wc -l | awk '{$1=$1};1' || echo -e "${RED}Error${NC}")
                if [[ $pruned_branches == "Error" ]]; then
                    pruned_branches="${RED}Error${NC}"
                else
                    git branch --merged main | grep -v '^\* main$' | xargs -n 1 git branch -d >/dev/null 2>&1
                fi
                branches=$(git branch | wc -l | awk '{$1=$1};1')

                # Output
                if [[ -z $changes && $branches -eq 1 && $current_branch == "main" && $pruned_branches == 0 ]]; then
                    echo -e "${GREEN}$dir: ${NC}"
                else
                    echo -n "${RED}$dir: ${NC}"
                    if [[ -n $changes ]]; then
                        echo -n "${BLUE}Non Committed Changes:${NC} $changes "
                    fi
                    if [[ $branches -ne 1 ]]; then
                        echo -n "${YELLOW} Too many branches: $branches${NC}"
                    fi
                    if [[ $current_branch != "main" ]]; then
                        echo -n "${BLUE} Current Branch: $current_branch${NC}"
                    fi
                    if [[ $pruned_branches -ne 0 ]]; then
                        echo -n "${YELLOW} Pruned Branches: $pruned_branches${NC}"
                    fi
                    echo "" # Move to the next line for the next repository
                fi
            ) &
        fi
    done < <(find "$target_dir" "${find_options[@]}" -type d -print0)

    wait # Wait for all background processes to finish
}
