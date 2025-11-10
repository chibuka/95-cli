# Deployment Guide

This guide explains how to deploy the 95 CLI so users can install it with `curl -fsSL https://ninefive.dev/install.sh | bash`

## Prerequisites

1. GitHub repository for the CLI
2. Domain name (e.g., ninefive.dev)
3. Web hosting to serve install.sh

## Step-by-Step Deployment

### 1. Update Repository URLs

Before deploying, update these files with your actual values:

**Cargo.toml:**
```toml
repository = "https://github.com/YOUR_USERNAME/ninefive-cli"
homepage = "https://ninefive.dev"
authors = ["Your Name <your.email@example.com>"]
```

**install.sh:**
```bash
REPO="YOUR_USERNAME/ninefive-cli"
```

### 2. Create First Release

Push your code to GitHub and create a release tag:

```bash
# Commit all changes
git add .
git commit -m "feat: initial release"
git push origin main

# Create and push a version tag
git tag v0.1.0
git push origin v0.1.0
```

GitHub Actions will automatically:
- Build binaries for all platforms (macOS Intel, macOS ARM, Linux x64, Linux ARM)
- Create a GitHub release
- Upload all binaries to the release

### 3. Host install.sh Script

You need to make `install.sh` accessible at `https://ninefive.dev/install.sh`

**Option A: GitHub Pages (Free)**

1. Create a new repo called `ninefive.dev` or use your existing website repo
2. Add `install.sh` to the repo
3. Enable GitHub Pages in repo settings
4. Point your domain to GitHub Pages

**Option B: Your Own Server**

1. Copy `install.sh` to your web server
2. Make it accessible at `https://yourdomain.com/install.sh`
3. Ensure proper CORS headers

**Option C: GitHub Raw (Quick Test)**

For testing, you can use:
```bash
curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/ninefive-cli/main/install.sh | bash
```

### 4. Test Installation

Test the installation on different platforms:

```bash
# Remove any existing installation
rm ~/.local/bin/95

# Test install script
curl -fsSL https://ninefive.dev/install.sh | bash

# Verify
95 --version
```

### 5. Future Releases

To release a new version:

```bash
# Update version in Cargo.toml
# Then commit and tag
git add Cargo.toml
git commit -m "chore: bump version to 0.2.0"
git push

git tag v0.2.0
git push origin v0.2.0
```

GitHub Actions will automatically build and release.

## DNS Configuration

Point your domain to the hosting service:

**For GitHub Pages:**
- Add CNAME record: `ninefive.dev` → `yourusername.github.io`
- Or A record to GitHub's IPs

**For your server:**
- Add A record pointing to your server IP

## Security Considerations

1. **HTTPS Required**: Always use HTTPS for install scripts
2. **Script Integrity**: Users should verify the script before piping to bash
3. **Checksums**: Consider adding SHA256 verification to install.sh
4. **Signed Releases**: Consider signing releases with GPG

## Monitoring

Monitor your releases:
- GitHub Actions: Check build success
- Download stats: GitHub Releases insights
- Error reports: GitHub Issues

## Marketing Your CLI

Once deployed, promote it:

1. Add to your website homepage
2. Update README with install command
3. Add badge to README: ![Install](https://img.shields.io/badge/install-curl%20%7C%20bash-blue)
4. Share on social media
5. Add to package manager directories (brew, cargo, etc.)

## Troubleshooting Deployment

**Build fails on GitHub Actions:**
- Check workflow logs
- Ensure Cargo.toml is valid
- Verify all dependencies compile on all platforms

**Install script 404:**
- Check DNS propagation
- Verify file is served correctly
- Check web server logs

**Binary not executable:**
- GitHub Actions should set `chmod +x`
- Check release workflow

## Next Steps

After successful deployment:

1. ✅ Add Homebrew formula (for `brew install 95`)
2. ✅ Publish to crates.io (for `cargo install ninefive-cli`)
3. ✅ Add auto-update functionality
4. ✅ Add analytics/telemetry (opt-in)
5. ✅ Create installer for Windows

## Support

For deployment issues, open an issue on GitHub or contact support.
