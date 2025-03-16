# Contributing Guidelines

Thank you for your interest in contributing to iam-manager! This document outlines the process for contributing to the project and how to get started.

## How to report a bug
* Open an issue at https://github.com/keikoproj/iam-manager
* What did you do? (how to reproduce)
* What did you see? (include logs and screenshots as appropriate)
* What did you expect?

## How to contribute a bug fix
* Open an issue at https://github.com/keikoproj/iam-manager and discuss it.
* Create a pull request for your fix.

## How to suggest a new feature
* Open an issue at https://github.com/keikoproj/iam-manager and discuss it.

## Developer Certificate of Origin (DCO)

The iam-manager project requires all contributors to certify that they can contribute under the terms of the [Developer Certificate of Origin (DCO)](https://developercertificate.org/). The DCO is a simple attestation that you are the creator of your contribution, and that you're allowed to submit it under the open source license used by this project.

### DCO Sign-Off Methods

The preferred way to certify the DCO is to add a line to every git commit message:

```
Signed-off-by: Jane Doe <jane.doe@example.com>
```

You must use your real name (no pseudonyms or anonymous contributions) and an email address that you actively use. You can automatically add this signature by using `git commit -s` whenever you commit.

### Setting Up Your Development Environment

To ensure all your commits are automatically signed and include the DCO sign-off:

1. Configure Git for your local repository:
   ```bash
   # Enable signing for all commits in this repository
   git config commit.gpgsign true
   
   # Add DCO sign-off trailer
   git config trailer.sign.key "Signed-off-by: "
   git config trailer.sign.ifmissing add
   git config trailer.sign.command 'echo "$(git config user.name) <$(git config user.email)>"'
   ```

2. For global configuration (apply to all repos):
   ```bash
   git config --global commit.gpgsign true
   git config --global trailer.sign.key "Signed-off-by: "
   git config --global trailer.sign.ifmissing add
   git config --global trailer.sign.command 'echo "$(git config user.name) <$(git config user.email)>"'
   ```

## Pull Request Process

1. Fork the repository and create your branch from `master`.
2. Make your changes and ensure all tests pass.
3. Update documentation for any new features or changes.
4. Ensure your commits include the DCO sign-off.
5. Submit a pull request.

## Code of Conduct

Please follow the [Contributor Covenant](https://www.contributor-covenant.org/version/2/0/code_of_conduct/) when participating in this project.

## License

By contributing to iam-manager, you agree that your contributions will be licensed under the project's [Apache 2.0 License](LICENSE).
