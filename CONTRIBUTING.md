# Contributing to iam-manager

Thank you for considering contributing to iam-manager! This document outlines the process for contributing to the project and provides guidelines to help you get started.

## Code of Conduct

By participating in this project, you agree to abide by the [Code of Conduct](.github/CODE_OF_CONDUCT.md). Please read it to understand the expectations for all interactions within the community.

## Ways to Contribute

There are many ways to contribute to iam-manager:

- **Report bugs**: Submit issues for any bugs you encounter
- **Suggest enhancements**: Submit ideas for new features or improvements
- **Improve documentation**: Help us improve or correct documentation
- **Submit code changes**: Contribute bug fixes or new features
- **Review code**: Help review pull requests from other contributors

## Development Workflow

### Setting up your development environment

1. Fork the repository on GitHub
2. Clone your fork locally: `git clone https://github.com/YOUR-USERNAME/iam-manager.git`
3. Add the upstream repository: `git remote add upstream https://github.com/keikoproj/iam-manager.git`
4. Create a new branch for your changes: `git checkout -b feature/your-feature-name`

### Making changes

1. Make your changes to the codebase
2. Write or update tests for the changes you make
3. Make sure your code passes all tests
4. Update documentation as needed
5. Commit your changes (see DCO section below)

## Pull Request Process

1. Push your changes to your fork: `git push origin feature/your-feature-name`
2. Open a pull request against the main repository
3. Ensure the PR description clearly describes the problem and solution
4. Include issue numbers if applicable (e.g., "Fixes #123")
5. Wait for maintainers to review your PR
6. Make any requested changes
7. Once approved, your PR will be merged

## Developer Certificate of Origin (DCO) Signing

We require all contributors to sign their commits with a Developer Certificate of Origin (DCO). The DCO is a lightweight way for contributors to certify that they wrote or otherwise have the right to submit the code they are contributing.

### What is DCO?

The DCO is a simple statement that you, as a contributor, have the legal right to make the contribution and agree to do so under the project's license. The full text of the DCO can be found at [developercertificate.org](https://developercertificate.org/).

### How to sign your commits with DCO

Add a `Signed-off-by` line to your commit messages using the `-s` flag:

```bash
git commit -s -m "Your commit message"
```

This will add a signature line to your commit message:

```
Signed-off-by: Your Name <your.email@example.com>
```

Make sure the email address used matches your GitHub account's email address.

### DCO Verification

Pull requests are automatically checked for DCO signatures. PRs that don't have properly signed commits will need to be fixed before they can be merged.

## Reporting Bugs

When reporting bugs, please include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Any relevant logs or screenshots
- Your environment information (OS, version, etc.)

Submit bug reports at: https://github.com/keikoproj/iam-manager/issues

## Suggesting Enhancements

When suggesting enhancements, please include:

- A clear and descriptive title
- A detailed description of the proposed functionality
- Rationale: why this enhancement would be valuable
- If possible, example use cases or implementations

Submit enhancement suggestions at: https://github.com/keikoproj/iam-manager/issues

## Communication

- GitHub Issues: For bug reports, feature requests, and project discussions
- Pull Requests: For code reviews and submitting changes
- Slack: Join our [Slack channel](https://keikoproj.slack.com/messages/iam-manager) for quick questions and community discussion

## License

By contributing to iam-manager, you agree that your contributions will be licensed under the project's [license](LICENSE).
