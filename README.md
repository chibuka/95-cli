# 95 CLI

```
 █████╗ ███████╗
██╔══██╗██╔════╝
╚██████║███████╗
 ╚═══██║╚════██║
 █████╔╝███████║
 ╚════╝ ╚══════╝
```

**Build your coding skills, one challenge at a time**

A command-line tool for practicing coding challenges with real-time validation and progress tracking. Your code runs locally, gets validated server-side, and tracks your progress as you level up your skills.

## Features

- **GitHub OAuth Authentication** - Secure login with your GitHub account
- **Local Execution** - Your code runs on your machine for fast feedback
- **Server-Side Validation** - Expected outputs never leave the server (prevents cheating)
- **Progress Tracking** - Automatically saves your progress as you complete stages
- **Cascading Tests** - Tests all prerequisite stages to ensure backward compatibility

## Installation

### Quick Install (Recommended)

**macOS/Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/chibuka/95-cli/main/install.sh | bash
```

### Manual Installation

1. Download the latest binary for your platform from [Releases](https://github.com/chibuka/95-cli/releases)
2. Extract and move to your PATH:
   ```bash
   # macOS/Linux
   sudo mv 95 /usr/local/bin/
   chmod +x /usr/local/bin/95
   ```

### Build from Source

Requirements: Go 1.23+

```bash
git clone https://github.com/chibuka/95-cli.git
cd 95-cli
go build -o 95
sudo mv 95 /usr/local/bin/
```

## Quick Start

### 1. Authenticate

```bash
95 login
```

This opens your browser for GitHub OAuth authentication. Your tokens are securely stored locally.

### 2. Initialize Your Project

```bash
95 init --cmd "python main.py"
```

This creates a `.95.yaml` config file with your project settings:
- Command to run your code
- Programming language
- Working directory (optional)

### 3. Test Locally

```bash
95 test <stage-uuid>
```

Run tests locally without submitting to the server. Great for debugging!

### 4. Submit Your Solution

```bash
95 run <stage-uuid>
```

Runs all tests and submits results to the server for validation.

## Commands

### `95 login`
Authenticate with GitHub OAuth. Opens your browser for authentication.

### `95 init`
Initialize project configuration.

**Flags:**
- `--cmd` - Command to run your code (required)
- `--lang` - Programming language (auto-detected if not specified)
- `--dir` - Working directory (default: current directory)

**Example:**
```bash
95 init --cmd "cargo run" --lang rust
95 init --cmd "node index.js" --lang javascript
95 init --cmd "python main.py" --lang python
```

### `95 test <stage-uuid>`
Run tests locally without submitting to the server.

**Example:**
```bash
95 test a85fdf04-a98e-4747-aa38-6e38babe663c
```

### `95 run <stage-uuid>`
Run all tests and submit results for validation.

**Example:**
```bash
95 run a85fdf04-a98e-4747-aa38-6e38babe663c
```

**Cascading Tests:**
When you run a stage, 95 CLI automatically tests all prerequisite stages (e.g., running stage 5 will test stages 1-5). This ensures your solution doesn't break previous functionality.

### `95 logout`
Clear stored credentials and log out.

## Configuration

### Project Configuration (`.95.yaml`)

Created by `95 init` in your project directory:

```yaml
cmd: python main.py
language: python
workingDir: .
```

### User Credentials (`~/.95cli/config.json`)

Created automatically during login. Contains:
- Access token (JWT, expires in 24 hours)
- Refresh token (valid for 30 days)
- User information (ID, username, email)
- API URL

**Security:**
- File permissions set to `0600` (owner read/write only)
- Tokens are automatically refreshed when expired
- Never committed to version control

## Example Session

```bash
# First time setup
$ 95 login
 █████╗ ███████╗
██╔══██╗██╔════╝
╚██████║███████╗
 ╚═══██║╚════██║
 █████╔╝███████║
 ╚════╝ ╚══════╝

Build your coding skills, one challenge at a time

Opening browser for GitHub authentication...
✓ Logged in successfully!

# Initialize your project
$ 95 init --cmd "python main.py"
✓ Config saved to .95.yaml

# Test locally
$ 95 test a85fdf04-a98e-4747-aa38-6e38babe663c
Stage 01: Handle echo command
├─ Echo simple string
     $ echo hello
       hello

├─ Echo email address
     $ echo test@example.com
       test@example.com

└─ Echo numbers
     $ echo 123
       123

# Submit solution
$ 95 run a85fdf04-a98e-4747-aa38-6e38babe663c
Stage 01: Handle echo command
  ├─ ✓ Echo simple string
  ├─ ✓ Echo email address
  └─ ✓ Echo numbers

✓ All 3 tests passed!

→ Check your browser for live progress updates and stage completion!
```

## How It Works

### Authentication Flow
1. CLI starts local HTTP server on port 9417
2. Opens browser to GitHub OAuth page
3. User authenticates with GitHub
4. Backend redirects to CLI's local server with JWT tokens
5. CLI saves tokens securely and closes server

### Test Execution Flow
1. CLI fetches test configuration from backend (assertions stripped)
2. CLI runs your code locally with test inputs
3. CLI captures stdout, stderr, exit codes, and HTTP responses
4. CLI submits results to backend for validation
5. Backend validates against server-side assertions
6. Backend returns pass/fail with detailed feedback

### Cascading Tests
When you request stage N, the backend returns tests for stages 1..N:
- Ensures backward compatibility
- Prevents regressions
- Validates entire solution path


### Release Process

Releases are automated via GitHub Actions.


## Troubleshooting

### "Authentication failed"
Run `95 login` to re-authenticate. Your session may have expired.

### "Stage not found"
Double-check the stage UUID. You can find it on the 95 website.

### "Command not found: 95"
Make sure `~/.local/bin` is in your PATH:
```bash
export PATH="$PATH:$HOME/.local/bin"
```

### Tests fail locally but should pass
Run `95 test` to see detailed output and debug your solution.

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
5. Submit a pr, thank you :)

## License

MIT License - see [LICENSE](LICENSE) for details

## Links

- **Website:** https://95ninefive.dev
- **GitHub:** https://github.com/chibuka/95-cli
- **Personal GitHub:** https://github.com/grainme

---

ありがとうございます 💯
