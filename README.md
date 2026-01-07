# 95 CLI

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

Requires Go 1.23+:

```bash
go install github.com/chibuka/95@latest
```

Make sure `$GOPATH/bin` or `$HOME/go/bin` is in your PATH:

```bash
# macOS/Linux
export PATH=$PATH:$HOME/go/bin
```

Verify installation:

```bash
95 --version
```

### Build from Source (Alternative)

```bash
git clone https://github.com/chibuka/95.git
cd 95
go build -o 95
sudo mv 95 /usr/local/bin/
```

## Quick Start

### 1. Authenticate

```bash
95 login
```

Opens your browser for GitHub OAuth authentication. Tokens are stored securely locally.

### 2. Initialize Your Project

```bash
95 init --cmd "python main.py" // in case you're using python
```

Creates a `config.json` file with your project settings. The language is automatically detected from your command.

### 3. Test Locally

```bash
95 test <stage-uuid>
```

Run tests locally without submitting to the server. Useful for debugging.

### 4. Submit Your Solution

```bash
95 run <stage-uuid>
```

Runs all tests and submits results to the server for validation. Cascading tests ensure previous stages are also validated.

---

## Commands

- `95 login` — Authenticate with GitHub OAuth  
- `95 logout` — Clear credentials and log out  
- `95 init` — Initialize project configuration (`--cmd` or positional argument)  
- `95 test <stage-uuid>` — Run tests locally  
- `95 run <stage-uuid>` — Run all tests and submit results  

---

## Configuration

### Project Configuration (`config.json`)

Created by `95 init` in your project directory:

```json
{
  "runCommand": "python main.py",
  "language": "PYTHON"
}
```

**Note:** Language is automatically detected from your run command. Supported languages include Python, Go, Java, Rust, JavaScript (Node.js), C, and C++.

### User Credentials (`~/.95cli/config.json`)

Contains:
- Access token (JWT, expires in 24 hours)
- Refresh token (valid for 30 days)
- User info (ID, username)
- API URL  

**Security:** Permissions `0600`, auto-refresh, never committed.

---

## Example Session

```bash
# Login
$ 95 login
 █████╗ ███████╗
██╔══██╗██╔════╝
╚██████║███████╗
 ╚═══██║╚════██║
 █████╔╝███████║
 ╚════╝ ╚══════╝

Build your coding skills, one challenge at a time

✓ Logged in successfully!

# Initialize project
$ 95 init --cmd "python main.py"
✓ Project initialized!
  Run command: python main.py
  Language: PYTHON

# Test stage locally
$ 95 test a85fdf04-a98e-4747-aa38-6e38babe663c

# Submit stage
$ 95 run a85fdf04-a98e-4747-aa38-6e38babe663c
✓ All tests passed!
```

---

## How It Works

### Authentication Flow
1. CLI starts a local HTTP server  
2. Opens browser for GitHub OAuth  
3. CLI receives tokens from the redirect and stores them securely  

### Test Execution Flow
1. Fetches test configuration from backend  
2. Runs code locally and captures output  
3. Submits results to backend for validation  
4. Returns pass/fail with detailed feedback  

### Cascading Tests
Running stage N tests all previous stages to ensure backward compatibility.

---

## Troubleshooting

- **"Command not found: 95"** — Ensure `$HOME/go/bin` is in your PATH
- **"Authentication failed"** — Run `95 login` again (your session may have expired)
- **"Stage not found"** — Verify stage UUID

---

## Contributing

1. Fork the repository  
2. Create a feature branch  
3. Make changes  
4. Submit a PR  

---

## License

MIT License - see [LICENSE](LICENSE)

## Links

- **Website:** https://95ninefive.dev  
- **GitHub:** https://github.com/chibuka/95  
- **Personal GitHub:** https://github.com/grainme

---

ありがとうございます 💯
