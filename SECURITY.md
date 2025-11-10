## What Data Does 95 CLI Access?

**Data We Send:**
- Your code submissions (sent to api.ninefive.dev)
- Submission codes (JWT tokens from ninefive.dev)

**Data Stored Locally:**
- None. The CLI is stateless and stores nothing on your machine.

**Data We Don't Access:**
- Files outside your submission directory
- Environment variables
- Browser data or cookies
- System credentials
- Any personal information

## How It Works

1. You provide a submission code (JWT token) from ninefive.dev
2. CLI creates a ZIP archive of your project directory
3. ZIP is uploaded to api.ninefive.dev via HTTPS
4. Results are displayed in your terminal

**Same Backend as Web Interface:**
The CLI uses the exact same API as the ninefive.dev website. If you trust the website, you can trust the CLI.

## What We Don't Do

- No background processes
- No telemetry or analytics
- No auto-updates without consent
- No third-party data sharing
- No elevated/sudo permissions required
- No persistent authentication storage

## Files Excluded from Submissions

The CLI automatically excludes sensitive and unnecessary files:
- `.git` directory
- `node_modules`, `target`, `__pycache__` (dependencies/build artifacts)
- `.env`, `.env.local` (environment files)
- IDE configs (`.vscode`, `.idea`)
- System files (`.DS_Store`, `Thumbs.db`)

## Transparency

**Verbose Mode:**
Use `--verbose` flag to see exactly what the CLI is doing:
```bash
95 submit [submission_code] --verbose
```

This shows:
- API endpoints being called
- Request/response data
- JWT token contents (non-sensitive parts)

**Open Source:**
This project is open source. You can inspect the code at:
https://github.com/chibuka/95-cli

## Reporting Security Issues

If you discover a security vulnerability, please email:
**boufaroujmarouan@gmail.com**

## Best Practices

**Protect Your Submission Code:**
- Treat submission codes like passwords
- Don't share them publicly
- They expire after 7 days
- Each challenge/language has a unique code

**Review Before Submitting:**
- Use `--verbose` to see what's being uploaded
- Check your directory doesn't contain secrets
- Ensure `.env` files are excluded (they are by default)