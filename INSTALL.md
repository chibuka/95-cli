# Installation Guide

## Quick Install (Recommended)

Install 95 CLI with a single command:

```bash
curl -fsSL https://95ninefive.dev/install.sh | bash
```

After installation, reload your shell or run:
```bash
export PATH="$PATH:$HOME/.local/bin"
```

## Verify Installation

```bash
95 --version
```

## Usage

Run the CLI in your project directory:

```bash
95
```

Or specify a submission code directly:

```bash
95 submit <submission-code>
```

Or specify a custom path:

```bash
95 submit <submission-code> --path /path/to/project
```

## Supported Platforms

- ✅ macOS (Intel & Apple Silicon)
- ✅ Linux (x86_64 & ARM64)

## Manual Installation

If you prefer to install manually:

1. Download the binary for your platform from [GitHub Releases](https://github.com/chibuka/95-cli/releases)
2. Make it executable: `chmod +x 95-<platform>`
3. Move to PATH: `mv 95-<platform> ~/.local/bin/95`

## Updating

Re-run the install script to update to the latest version:

```bash
curl -fsSL https://ninefive.dev/install.sh | bash
```

## Uninstalling

Remove the binary:

```bash
rm ~/.local/bin/95
```

## Troubleshooting

### Command not found

Make sure `~/.local/bin` is in your PATH:

```bash
echo 'export PATH="$PATH:$HOME/.local/bin"' >> ~/.bashrc  # or ~/.zshrc
source ~/.bashrc  # or ~/.zshrc
```

### Permission denied

The install script needs to write to `~/.local/bin`. Make sure the directory exists and is writable:

```bash
mkdir -p ~/.local/bin
chmod u+w ~/.local/bin
```

### Download failed

Check your internet connection and try again. If the problem persists, try manual installation.

## Support

For issues, visit: https://github.com/chibuka/95-cli/issues
