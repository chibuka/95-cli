# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2025-11-10

### Added
- Initial release of ninefive CLI
- `submit` command for uploading code to ninefive.dev
- Automatic ZIP creation with smart file exclusions
- Real-time submission feedback
- JWT-based authentication via submission codes
- Verbose mode (`--verbose`) for debugging and transparency
- Terminal UI with spinners and progress indicators
- Automatic exclusion of common build artifacts and sensitive files:
  - `.git`, `node_modules`, `target`, `__pycache__`
  - `.env` files
  - IDE configs (`.vscode`, `.idea`)
  - System files (`.DS_Store`, etc.)
- Cross-platform support:
  - macOS (Intel and Apple Silicon)
  - Linux (x86_64 and ARM64)
- Interactive submission code prompt
- Submission details table display
- Custom path support (`--path` flag)

> Open source for transparency

[Unreleased]: https://github.com/grainme/ninefive-cli/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/grainme/ninefive-cli/releases/tag/v0.1.0
