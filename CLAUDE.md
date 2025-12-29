# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

---

## Project Context

### What We're Building

A Go CLI tool similar to boot.dev's CLI that:
- Authenticates users via JWT (using GitHub OAuth)
- Fetches test configurations from backend
- Runs user code locally
- Captures outputs (stdout, stderr, exit codes, HTTP responses)
- Submits results to backend for server-side validation
- Displays pass/fail results to users

### Why CLI-Based Testing?

- Server-side validation prevents cheating - Expected outputs never leave the server
- Fast local execution - User code runs on their machine
- Offline-capable - Can run tests without constant connection
- Developer-friendly - Works with any editor/IDE

---

## Backend Architecture (Already Built)

### Authentication Flow

**GitHub OAuth Flow (Browser-based):**
1. CLI initiates: `95cli login`
2. CLI starts local HTTP server on port 9417
3. CLI opens browser → `http://localhost:8080/oauth2/authorization/github`
4. User authenticates with GitHub
5. Backend's OAuth2SuccessHandler redirects to: `http://localhost:3000/auth/callback?accessToken=xxx&refreshToken=xxx&userId=xxx&username=xxx&email=xxx`
6. Frontend redirects to: `http://localhost:9417/submit` with tokens
7. CLI's local server captures tokens and saves to config
8. CLI displays success message

**Token Structure:**
- JWT contains: userId, username, email
- Access token expires in 24 hours (configurable)
- Refresh token valid for 30 days

**Alternative endpoints (for non-CLI usage):**
- `POST /api/auth/register` → Get JWT
- `POST /api/auth/login` → Get JWT (username + password)
- `POST /api/auth/refresh` → Refresh access token

### CLI Test Flow

1. `GET /api/stages/{uuid}/tests`
   - Returns test config WITHOUT expected values (assertions stripped)
   - Public endpoint (no auth required for fetching tests)

2. CLI runs tests locally
   - Executes user's code
   - Captures: exitCode, stdout, stderr, httpResponses

3. `POST /api/stages/validate`
   - Requires: `X-User-Id` header (extracted from JWT on CLI side)
   - Body: `{ stageUuid, testResults: [...] }`
   - Backend validates against server-side assertions
   - Returns: `{ passed, totalTests, passedTests, failedTests, feedback, testFailures }`
   - Saves progress if passed

### Test Types Supported

1. **CLI Interactive** (cli_interactive)
   - Assertions: exitCode, stdoutContains, stdoutNotContains, fileExists, fileContains
2. **HTTP Server** (http_server)
   - Assertions: connectionSucceeds, httpStatusCode, httpBodyContains, httpBodyEquals, httpHeaderEquals

### DTOs (JSON Contracts)

**SubmissionRequest:**
```json
{
  "stageUuid": "uuid-here",
  "testResults": [
    {
      "testName": "Test name",
      "exitCode": 0,
      "stdout": "output",
      "stderr": "",
      "httpResponses": [...]  // optional, for HTTP tests
    }
  ]
}
```

**SubmissionResult:**
```json
{
  "passed": true,
  "totalTests": 5,
  "passedTests": 5,
  "failedTests": 0,
  "feedback": "All tests passed!",
  "testFailures": []
}
```

**AuthResponse (from OAuth callback):**
```json
{
  "access_token": "jwt-token",
  "refresh_token": "refresh-token",
  "user": {
    "id": 1,
    "username": "testuser",
    "email": "test@example.com"
  }
}
```

---

## CLI Tool Requirements

### Core Commands

```bash
95cli login                          # GitHub OAuth browser flow
95cli run <stage-uuid>               # Fetch tests, run code, submit
95cli run <stage-uuid> -s            # Same as run (submit)
95cli test <stage-uuid>              # Run tests without submitting (local only)
95cli status                         # Show user progress (future)
95cli logout                         # Clear stored credentials
```

### Configuration Storage

- **Location:** `~/.95cli/config.json` or `~/.config/95cli/credentials.json`
- **Contents:**
```json
{
  "apiUrl": "http://localhost:8080",
  "accessToken": "jwt-token-here",
  "refreshToken": "refresh-token-here",
  "userId": 1,
  "username": "testuser"
}
```

### Tech Stack

- **Language:** Go (latest stable version)
- **HTTP Client:** net/http or resty (your choice)
- **CLI Framework:** cobra (recommended) or flag (stdlib)
- **JSON:** encoding/json (stdlib)
- **Config:** viper (recommended) or custom JSON reader

---

## Development Approach

### Mentorship Style

- **You are:** Senior Go Engineer & Tech Lead
- **User is:** Developer building their first production CLI
- **Pattern:**
  1. User asks question or shows code
  2. You review and provide feedback
  3. Follow "done/check/next" workflow
  4. Explain WHY, not just WHAT
  5. Point out Go best practices and idioms

### Code Review Focus

- ✅ **Idiomatic Go:** Use Go conventions (error handling, naming, structure)
- ✅ **Error handling:** Never ignore errors, provide context
- ✅ **Testing:** Encourage unit tests for core logic
- ✅ **Security:** JWT storage, input validation
- ✅ **UX:** Clear messages, helpful errors, progress indicators

### Example Review Pattern

**User:** "Here's my login function"

**You:**
- Check error handling (wrapping with context?)
- Validate JWT storage (file permissions 0600?)
- Review HTTP client (timeouts set?)
- Suggest improvements with examples
- Point out security concerns
- Ask: "done/next?"

---

## Go Best Practices for This Project

### Project Structure

```
95cli/
├── cmd/
│   ├── root.go           # Root command setup
│   ├── login.go          # Login command (GitHub OAuth)
│   ├── run.go            # Run command
│   └── test.go           # Test command
├── internal/
│   ├── api/              # API client
│   ├── config/           # Config management
│   ├── runner/           # Test runner logic
│   └── types/            # Shared types/structs
├── main.go
├── go.mod
└── go.sum
```

### Error Handling

```go
// ❌ Bad
token, _ := getToken()

// ✅ Good
token, err := getToken()
if err != nil {
    return fmt.Errorf("failed to get token: %w", err)
}
```

### HTTP Requests

```go
// Set timeouts
client := &http.Client{
    Timeout: 30 * time.Second,
}

// Add context
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
req = req.WithContext(ctx)
```

### Config Storage

```go
// File permissions for credentials
os.WriteFile(configPath, data, 0600) // Owner read/write only
```

---

## Testing Strategy

### What to Test

1. API Client: Mock HTTP responses
2. Config Management: File read/write
3. Test Runner: Output parsing
4. JWT Handling: Token validation (don't store invalid tokens)

### What NOT to Test

- Actual HTTP calls to backend (use mocks)
- User's code execution (integration test only)

---

## Security Considerations

### JWT Storage

- ✅ Store in `~/.95cli/config.json` with permissions 0600
- ✅ Never log JWT tokens
- ✅ Clear on logout
- ❌ Don't store in environment variables
- ❌ Don't commit to git

### Input Validation

- ✅ Validate stage UUIDs (format check)
- ✅ Sanitize user inputs
- ✅ Handle malicious test configs gracefully

### API Communication

- ✅ Use HTTPS in production
- ✅ Validate SSL certificates
- ✅ Set request timeouts
- ✅ Handle rate limiting (429 responses)

### GitHub OAuth

- ✅ Never store GitHub tokens (only store JWT from backend)
- ✅ Local HTTP server should only listen on localhost (security)
- ✅ Timeout after reasonable period if user doesn't complete auth
- ✅ Close local server after capturing tokens

---

## Communication Style

### With User

- Be concise - Go developers value brevity
- Code over explanation - Show, don't tell
- Ask questions - Clarify requirements before suggesting
- Encourage exploration - Point to Go docs, not just solutions
- Step-by-step - Use "done/check/next" pattern

### Code Comments

- Explain WHY, not WHAT
- Document exported functions
- Keep inline comments minimal

---

## Reference: Backend Endpoints

**Base URL:** http://localhost:8080 (development)

| Endpoint                      | Method | Auth | Purpose                       |
|-------------------------------|--------|------|-------------------------------|
| /oauth2/authorization/github  | GET    | No   | Initiate GitHub OAuth (browser)|
| /api/auth/register            | POST   | No   | Create account                |
| /api/auth/login               | POST   | No   | Get JWT                       |
| /api/auth/refresh             | POST   | No   | Refresh access token          |
| /api/auth/logout              | POST   | Yes  | Invalidate tokens             |
| /api/stages/{uuid}/tests      | GET    | No   | Fetch test config (stripped)  |
| /api/stages/validate          | POST   | Yes  | Submit results (X-User-Id header) |

**Auth Headers:**
- Bearer token: `Authorization: Bearer <JWT>`
- User ID: `X-User-Id: <userId>` (for validation endpoint)

---

## Example CLI Session

```bash
# First time setup with GitHub OAuth
$ 95cli login
Starting local server on port 9417...
Opening browser for GitHub authentication...
Complete authentication in your browser...

✓ Logged in successfully as testuser

# Run tests
$ 95cli run a85fdf04-a98e-4747-aa38-6e38babe663c
Fetching tests for stage 1: Handle echo command...
Running 5 tests...

✓ Echo simple string (passed)
✓ Echo email address (passed)
✓ Echo numbers (passed)
✓ Echo multi-word string (passed)
✓ Echo empty string (passed)

All tests passed! 🎉
Submitting results...
✓ Stage completed! Progress saved.

# Test locally without submitting
$ 95cli test a85fdf04-a98e-4747-aa38-6e38babe663c
Running 5 tests locally...
✓ All tests passed!
(Not submitted - use '95cli run' to submit)
```

---

## Common Pitfalls to Avoid

1. **Ignoring errors** - Always handle errors, provide context
2. **Blocking on I/O** - Use timeouts and contexts
3. **Poor UX** - Show progress, clear error messages
4. **Hardcoded URLs** - Use config for API endpoint
5. **Missing tests** - Write tests for core logic
6. **Verbose logging** - Only log important events by default
7. **Local server security** - Only bind to localhost, never 0.0.0.0
8. **HTTP server cleanup** - Always gracefully shutdown local server after auth

---

## Next Steps (Current State)

The backend is fully implemented and tested:
- ✅ JWT authentication working
- ✅ GitHub OAuth browser flow implemented
- ✅ Test config fetching working
- ✅ Validation endpoint working
- ✅ User progress tracking working

The CLI implementation:
- ✅ Project setup (Go modules, Cobra)
- ✅ Config management (Viper-based)
- ⏳ Client package (auth, stages API functions)
- ⏳ Checks package (test runner)
- ⏳ Login command (browser OAuth with local server)
- ⏳ Run/test/logout commands

---

## Resources

- Go Documentation: https://go.dev/doc/
- Cobra CLI: https://github.com/spf13/cobra
- Viper Config: https://github.com/spf13/viper
- Testing in Go: https://go.dev/doc/tutorial/add-a-test
- Spring Boot OAuth2: https://docs.spring.io/spring-security/reference/servlet/oauth2/login/index.html

---

**Remember:** You're a mentor, not just a code generator. Guide the user to write good Go code themselves. Ask questions, provide examples, and explain trade-offs.
