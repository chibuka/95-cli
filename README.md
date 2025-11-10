# ninefive CLI

Command-line tool for submitting code to the [ninefive](https://ninefive.dev) code challenge platform.

## Installation

### Quick Install (Recommended)

```bash
curl -fsSL https://95ninefive.dev/install.sh | bash
```

### From Source

```bash
# Clone the repository
git clone https://github.com/chibuka/95-cli.git
cd 95-cli

# Build release binary
cargo build --release

# The binary will be at target/release/95
```

### Manual Installation

Copy the binary to your PATH:

```bash
sudo cp target/release/95 /usr/local/bin/
```

Or add the binary directory to your PATH:

```bash
export PATH="$PATH:/path/to/ninefive-cli/target/release"
```

## Usage

### 1. Get Submission Code

1. Go to [ninefive.dev](https://ninefive.dev)
2. Login with GitHub
3. Navigate to a challenge (e.g., "build your own shell")
4. Select your language and stage
5. Copy the submission code shown in the "OR USE CLI" section

The code looks like: `95-eyJhbGciOi...`

### 2. Submit Your Code

Navigate to your project directory and run:

```bash
95 submit 95-eyJhbGciOi...
```

The CLI will:
1. Create a ZIP archive of your current directory
2. Upload it to ninefive
3. Show real-time test results (on your browser)

### Examples

**Submit current directory:**
```bash
95 submit 95-abc123xyz
```

**Submit a specific directory:**
```bash
95 submit 95-abc123xyz --path ./my-solution
```

**Show verbose output:**
```bash
95 submit 95-abc123xyz --verbose
```

## Why Use the CLI?

**Before (Manual ZIP):**
1. Edit code
2. Compress the project folder
3. Go to browser
4. Click upload
5. Select file
6. Click submit
7. Wait for results

> Next stage, you make a code change and you'll have to redo the same steps.

**After (CLI):**
1. Edit code
2. Run `95 submit <code>`
3. Wait for results (via browser)


## Troubleshooting

### "Invalid or expired submission code"

Your submission code has expired (valid for 7 days). Get a new code from the web interface.

### "API error: File size exceeds 50MB limit"

Your project is too large. Ensure you're not including build artifacts or dependencies.


## License
MIT
