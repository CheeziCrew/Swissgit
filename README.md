# Swiss Git

![Untitled_Artwork_15](https://github.com/CheeziCrew/Swissgit/assets/110965999/0edfe55f-38a2-4d06-9c39-5b60ff7f5441)

## Description

Swiss Git is a "comprehensive" tooling solution designed to streamline and simplify Git repository management, particularly when dealing with multiple repositories simultaneously.

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

- Go 1.23 or higher
- Git
- SSH key set up on GitHub
- GitHub personal access token

## Installation

### Using Released Binaries

You can download the latest binaries from the [Releases](https://github.com/CheeziCrew/swissgit/releases) section on GitHub.

1. Download the appropriate binary for your operating system.
2. Extract the binary and place it in a directory included in your system's PATH.

### Building from Source

1. Clone the repository:

   ```sh
   git clone https://github.com/CheeziCrew/swissgit.git
   cd swissgit
   ```

2. Build the project:

   ```sh
   go build
   ```

3. (Optional) Install the binary:
   ```sh
   go install
   ```

### Environment Setup

1. Create a `.env` file in the root of the project:

   ```sh
   touch .env
   ```

2. Add your GitHub personal access token and SSH key name to the `.env` file:

   ```env
   GITHUB_TOKEN=your_github_token
   SSH_KEY=your_ssh_key_name
   ```

   Example:

   ```env
   GITHUB_TOKEN=ghp_yourGitHubTokenHere
   SSH_KEY=id_ed25519
   ```

3. Ensure your SSH key is set up on GitHub. You can follow the instructions [here](https://docs.github.com/en/authentication/connecting-to-github-with-ssh).

## Usage

Swiss Git provides several commands to manage your repositories:

- **automerge**: Enable automerge for repository (currently requires github CLI)
- **status**: Check the status of repositories.
- **branches**: List local, remote, and stale branches in the repository.
- **clone**: Clone a repository or all repositories from a GitHub organization.
- **commit**: Add all files, commit changes, and push to the remote repository.
- **pullrequest**: Commit all changes and create a pull request on GitHub.
- **cleanup**: Reset changes, update the main branch, and prune branches.

### Example Commands

- Enable automerge:

  ```sh
  swissgit automerge -t "Your PR Title"
  ```

- Check the status of repositories:

  ```sh
  swissgit status --all --path /path/to/repos --verbose
  ```

- List branches:

  ```sh
  swissgit branches --path /path/to/repo
  ```

- Clone all repositories in a organisation:

  ```sh
  swissgit clone --org CheeziCrew --path /path/to/folder/
  ```

- Commit changes:

  ```sh
  swissgit commit --message "Your commit message"
  ```

- Create multiple pull request:

  ```sh
  swissgit pullrequest --all --message "Your PR title" --branch feature-branch
  ```

- Cleanup repositories:
  ```sh
  swissgit cleanup --path /path/to/repo --drop
  ```

## Features

- **Status Check**: Quickly check the status of multiple repositories.
- **Branch Management**: List and manage local and remote branches.
- **Repository Cloning**: Clone individual repositories or all repositories from an organization.
- **Commit and Push**: Easily commit changes and push to remote repositories.
- **Pull Requests**: Create pull requests directly from the command line.
- **Cleanup**: Reset changes, update branches, and prune old branches.

## Documentation

### Automerge Command
The `automerge` command enables automerge of a pull request in a repository. Currently relies on github CLI to perform the action

### Status Command

The `status` command checks the status of repositories. It can scan directories recursively if the `--all` flag is used.

### Branches Command

The `branches` command lists local, remote, and stale branches in the repository.

### Clone Command

The `clone` command clones a repository or all repositories from a GitHub organization.

### Commit Command

The `commit` command adds all files, commits changes, and pushes to the remote repository.

### Pull Request Command

The `pullrequest` command commits all changes and creates a pull request on GitHub.

### Cleanup Command

The `cleanup` command resets changes, updates the main branch, and prunes branches.

## Contributing

Feel free to create a pull request if you have suggestions on changes. Or create an issue if you find something that is behaving weirdly, have a question or suggestion.

## License

This project is licensed under the [MIT License](LICENSE). You can find the full text of the license in the [LICENSE](LICENSE) file.

## Acknowledgements

- **Theo the Cat**: for moral support

Contribute in any way, shape or form and your name might end up here.

## Contact Information

For questions or concerns, contact [swissgittools@cheezi.se](mailto:swissgittools@cheezi.se) or create an issue here on GitHub.

## FAQs

**Q: Where the duck is the .exe?**

A: Sorry we are developers here. We don't do .exes. Just kidding you find it under

**Q: Are all my questions answered here?**

A: Maybe. But feel free to ask anyway! We're here to help.
