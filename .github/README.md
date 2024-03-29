# Swiss Git

![Untitled_Artwork_15](https://github.com/CheeziCrew/Swissgit/assets/110965999/0edfe55f-38a2-4d06-9c39-5b60ff7f5441)

## Description

Swiss Git is a comprehensive tooling solution designed to streamline and simplify Git repository management, particularly when dealing with multiple repositories simultaneously.

## Table of Contents

1. [Description](#description)
2. [Table of Contents](#table-of-contents)
3. [Requirements](#requirements)
4. [Installation](#installation)
5. [Usage](#usage)
6. [Features](#features)
7. [Documentation](#documentation)
8. [Contributing](#contributing)
9. [License](#license)
10. [Acknowledgements](#acknowledgements)
11. [Contact Information](#contact-information)
12. [FAQs](#faqs)

## Requirements

Before installing Swiss Git, ensure you have the following requirements installed:

- **Git**: Swiss Git is built on top of Git. You must have Git installed on your system. You can download it from [git-scm.com](https://git-scm.com/).
- **GitHub CLI**: Some features of Swiss Git, such as cloning team repositories, require the GitHub CLI. You can install it by following the instructions on the [GitHub CLI documentation page](https://cli.github.com/manual/installation).

Please also note that `git config --global push.autoSetupRemote true` is expected to be set for the `commit` and `pullrequest` commands to work as intended.

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

     For bonus points, add the following lines to your Zsh configuration file to enable tab completion for Swiss Git commands:

     ```bash
     source /path/to/your/repository/completions/_swissgit
     autoload -Uz compinit && compinit
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

```
Usage: swissgit COMMAND

Commands:
  status                Recursively checks the status of repositories. If current dir is a git repo, it will only check that repo.
  branches              Recursively checks the branch status of repositories. If current dir is a git repo, it will only check that repo.
  cleanup               Clean up your repositories. Check out and update main, drop merged branches and drop no longer needed changes.
  clone                 Clone a teams repositories with SSH. Requires a personal access token.
  commit                Create and push a commit on the current branch or a new one. Without a PR
  pullrequest           Create a pull request. Creates a branch, commits all your changes, and creates a pull request.
  help                  Show this help message and exit
```

## Features

- Provides concise overview of Git repository status.
- Manages multiple Git repositories efficiently.
- Offers insights into the branches within Git repositories.
- Automates the process of creating pull requests in Git repositories.

## Documentation

### Branches

This command offers insights into the branches within Git repositories located in the current directory. It provides details on local and remote branches, highlighting the main branch and identifying stale branches for cleanup. Will only check the current directory if it is a git repo. Otherwise it will check all subdirectories.

#### Usage

```bash
swissgit branches [-h]
```

- `-h`: Show help message

#### Example Output

- \(Repo Name\): (\(L\)Local Branches/\(R\)Remote Branches): \(Main Branch\); \(Other Branches\); \(Stale Branches\)

##

### Cleanup

This bash script provides a streamlined way to manage multiple Git repositories within a specified directory. It checks for uncommitted changes, updates the main branch, prunes merged branches, and highlights potential issues like excessive local branches or unmerged changes.

#### Usage

```bash
swissgit cleanup [-h] [-d <drop_changes>] [-a <all_repos>] [-f <target_dir>]
```

- `-h`: Show help message
- `-a`: Apply changes to all repositories within the directory (recursive).
- `-d`: Drop all changes including untracked files.
- `-f <target_dir>`: Specify the directory path where the repositories are located. (Default: current directory)
- `-d`: Stash local changes before cleaning up.

#### Example Output

- Green (✓): Repository is clean and up-to-date.
- Yellow: Indicates potential issues such as uncommitted changes or excessive branches.
- Red: Indicates errors encountered during processing.

##

### Clone

This bash script automates the process of cloning all repositories associated with a team in an organization. It utilizes GitHub CLI and necessitates a personal access token for repository cloning.

#### Usage

```bash
swissgit Usage: swissgit clone -o <organization> -t <team> -k <token> [-f <target_dir>]
```

- `-h`: Show help message
- `-o <organization>`: Specify the GitHub organization."
- `-t <team>`: Specify the team within the organization."
- `-k <token>`: Specify the GitHub personal access token."
- `-f <target_dir>`: Specify the target directory for cloning. Defaults to the current directory.

#### Example Output

- Cloning repository: [repositoryName]
- Cloning completed successfully.

##

### Commit

This command wraps up the commands normally used for creating a commit and push it. It has the option to choose to commit to the current branch or providing a flag to create a new branch to commit to. This is useful for those times you don't want to create PR straight away or want to update a PR.

Note that the branch name will be added as a prefix to the commit message, so name your branches accordingly.

For example:

- Branch: `CHZ-001`
- Inputted commit message: `Implement feature and did stuff`
- Actual committed message: `CHZ-001: Implement feature and did stuff`

#### Usage

```bash
swissgit [-h] -c <commit_message> -b <branchname>"
```

- `-h`: Show help message
- `-a`: Apply changes to all repositories within the directory (recursive).
- `-b <branchname>`: Name of the new branch.
- `-c <commit_message>`: Commit message for the changes.
- `-n`: Do not push the commit to the remote repository.

##

### Pull Request

This command automates the process of creating pull requests in Git repositories. It simplifies branching, committing, pushing changes, and creating pull requests, either for a single repository or across multiple repositories.

Note that the branch name will be added as a prefix to the commit message, so name your branches accordingly.

For example:

- Branch: `CHZ-001`
- Inputted commit message: `Implement feature and did stuff`
- Actual committed message: `CHZ-001: Implement feature and did stuff`

#### Usage

```bash
swissgit pullrequest [-h] [-a] -b <branchname> -c <commit_message> [-p <pr_body>]
```

- `-h`: Show help message
- `-a`: Apply changes to all repositories within the current directory (optional).
- `-b <branchname>`: Name of the new branch.
- `-c <commit_message>`: Commit message for the changes.
- `[-p <pr_body>]`: Pull request body (optional).

#### Example Output

- Pull request created: [PR_URL]

##

### Status

This Bash script provides a concise overview of the status of Git repositories within the current directory. It displays information such as branch names, commits ahead/behind, and changes in a color-coded format. Will only check the current directory if it is a git repo. Otherwise it will check all subdirectories.

#### Usage

```bash
swissgit status [-h]
```

- `-h`: Show help message

#### Example Output

- Branch name: Green for 'main', Yellow for other branches.
- Commits Ahead/Behind: Green for ahead, Red for behind.
- Change Summary: Yellow for modified, Green for added, Red for deleted, Blue for untracked files.
- ✓: Indicates no changes and not ahead/behind.

## Contributing

Feel free to create a pull request if you have suggestions on changes. Or create an issue if you find something that is behaving weirdly, have a question or suggestion.

## License

This project is licensed under the [MIT License](LICENSE). You can find the full text of the license in the [LICENSE](LICENSE) file.

## Acknowledgements

- **Theo the Cat**: for moral support

Contribute in any way, shape or form and your name might end up here.

## Contact Information

For questions or concerns, contact [swissgittools@gmail.com](mailto:swissgittools@gmail.com) or create an issue here on GitHub.

## FAQs

**Q: Where the duck is the .exe?**

A: Sorry we are developers here. We don't do .exes.

**Q: Are all my questions answered here?**

A: Maybe. But feel free to ask anyway! We're here to help.
