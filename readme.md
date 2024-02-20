# Swiss Git

## Description

Swiss Git is a comprehensive tooling solution designed to streamline and simplify Git repository management, particularly when dealing with multiple repositories simultaneously.

## Table of Contents

1. [Description](#description)
2. [Table of Contents](#table-of-contents)
3. [Installation](#installation)
4. [Usage](#usage)
5. [Features](#features)
6. [Documentation](#documentation)
7. [Contributing](#contributing)
8. [License](#license)
9. [Acknowledgements](#acknowledgements)
10. [Contact Information](#contact-information)
11. [FAQs](#faqs)

## Installation

To install this repository, follow these steps:

1. Clone the Repository: Clone this repository to your local machine.

   ```bash
   git clone git@github.com:cheezi0747/Swissgit.git
   ```

2. Include `swissgit.sh` in Your Path:

   - For Zsh:

     Add the following line to your Zsh configuration file (e.g., `~/.zshrc`):

     ```bash
     source /path/to/your/repository/swissgit.sh
     ```

   - For Bash:

     Add the following line to your Bash configuration file (e.g., `~/.bashrc` or `~/.bash_profile`):

     ```bash
     source /path/to/your/repository/swissgit.sh
     ```

   - For Git Bash on Windows:

     Edit your `~/.bash_profile` file in your user directory (create it if it doesn't exist), and add the following line:

     ```bash
     source /c/path/to/your/repository/swissgit.sh
     ```

   Make sure to replace `/path/to/your/repository` with the actual path to the directory where you cloned the repository.

3. Reload Your Shell Configuration: After adding the sourcing line to your configuration file, reload your shell configuration or open a new terminal session to apply the changes.

## Usage

```bash
Usage: swissgit [-h | --help] [-s | --status] [-b | --branches] [-c | --clean [-a] [-b] [folder]]
[-p | --pullrequest [-a] <branchname> <commit_message> [PR_body]]

Options:
-h,     --help      Show this help message and exit
-s,     --status    Checks recursively the status of all repositories
-b,     --branches  Checks recursively the branch status of all repositories
-c,     --clean [-a] [-d] [folder]
                    Clean untracked files. Use -a to clean all, -d to drop local changes, and [folder] to specify a folder.
-p,     --pullrequest [-a] <branchname> <commit_message> [PR_body]
                    Create a pull request. Use -a for recursively doing for all subdirectories.
                    Creates a branch, commits all your changes and creates a pull pullrequest.
```

## Features

- Provides concise overview of Git repository status.
- Manages multiple Git repositories efficiently.
- Offers insights into the branches within Git repositories.
- Automates the process of creating pull requests in Git repositories.

## Documentation

### Cleanup `-c | --cleanup [-a] [-d] [folder]`

This bash script provides a streamlined way to manage multiple Git repositories within a specified directory. It checks for uncommitted changes, updates the main branch, prunes merged branches, and highlights potential issues like excessive local branches or unmerged changes.

#### Usage

```bash
swissgit -c [-d] [-a] [directory_path]
```

- \`-d\`: Drop all changes including untracked files.
- \`-a\`: Apply changes to all repositories within the directory (recursive).
- \`directory_path\`: Specify the directory path where the repositories are located. (Default: current directory)

#### Example Output

- Green (✓): Repository is clean and up-to-date.
- Yellow: Indicates potential issues such as uncommitted changes or excessive branches.
- Red: Indicates errors encountered during processing.

##

### Branches ` -b | --branches`

This command offers insights into the branches within Git repositories located in the current directory. It provides details on local and remote branches, highlighting the main branch and identifying stale branches for cleanup.

#### Usage

```bash
swissgit -b
```

#### Example Output

- \(Repo Name\): (\(L\)Local Branches/\(R\)Remote Branches): \(Main Branch\); \(Other Branches\); \(Stale Branches\)

##

### Status `-s | --status`

This Bash script provides a concise overview of the status of Git repositories within the current directory. It displays information such as branch names, commits ahead/behind, and changes in a color-coded format.

#### Usage

```bash
swissgit -s
```

#### Example Output

- Branch name: Green for 'main', Yellow for other branches.
- Commits Ahead/Behind: Green for ahead, Red for behind.
- Change Summary: Yellow for modified, Green for added, Red for deleted, Blue for untracked files.
- ✓: Indicates no changes and not ahead/behind.

##

### Pull Request `-p | --pullrequest`

This Bash script automates the process of creating pull requests in Git repositories. It simplifies branching, committing, pushing changes, and creating pull requests, either for a single repository or across multiple repositories.

#### Usage

```bash
. swissgit -p [-a] <branchname> <commit_message> [PR_body]
```

- `-a`: Apply changes to all repositories within the current directory (optional).
- `<branchname>`: Name of the new branch.
- `<commit_message>`: Commit message for the changes.
- `[PR_body]`: Pull request body (optional).

#### Example Output

- Pull request created: [PR_URL]

## Contributing

Explain how users can contribute to the project.

## License

Specify the project's license here.

## Acknowledgements

Thank contributors, libraries, or other resources here.

## Contact Information

## FAQs

Include frequently asked questions here.
