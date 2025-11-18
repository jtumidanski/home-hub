# GitHub Actions Workflows

This directory contains CI/CD workflows for the Home Hub monorepo with intelligent dependency-aware build detection.

## Overview

The workflows automatically detect which services need building based on file changes, with special handling for the shared Go library (`packages/shared-go`).

### Workflows

1. **`pr-build.yml`** - Pull Request Validation
2. **`main-publish.yml`** - Main Branch Publishing

---

## Pull Request Build (`pr-build.yml`)

**Purpose**: Validate that changed services build successfully without publishing images.

### Triggers
- Pull request opened, synchronized, or reopened against `main` or `master`
- Manual workflow dispatch

### Behavior
1. Detects which services changed
2. Builds only affected services
3. Does NOT push images to registry
4. Reports build status on PR

### Manual Trigger
You can manually trigger this workflow from the Actions tab:
- **Force build all**: Check this option to build all services regardless of changes

---

## Main Branch Publish (`main-publish.yml`)

**Purpose**: Build and publish Docker images to GitHub Container Registry (GHCR) when code is merged to main/master.

### Triggers
- Push to `main` or `master` branches
- Manual workflow dispatch

### Behavior
1. Detects which services changed
2. Builds affected services
3. Publishes images to GHCR with multiple tags:
   - `latest`: Always points to most recent main build
   - `sha-<short-sha>`: Specific commit reference
   - `v*`: Semantic version (if Git tag exists)
4. Generates summary with pull commands

### Manual Trigger
You can manually trigger this workflow from the Actions tab:
- **Services to publish**: Enter comma-separated service names (e.g., `svc-users,admin`) or leave empty for auto-detection

---

## Change Detection Logic

The `.github/actions/detect-changes` action implements smart dependency detection:

### Go Services Dependency Graph
```
packages/shared-go
  ├─→ svc-users
  ├─→ svc-tasks
  ├─→ svc-weather
  └─→ svc-reminders
```

### Detection Rules

| Changed Files | Services Built |
|---------------|----------------|
| `packages/shared-go/**` | ALL Go services (svc-users, svc-tasks, svc-weather, svc-reminders) |
| `go.work` or `go.work.sum` | ALL Go services |
| `apps/svc-users/**` | svc-users only |
| `apps/svc-tasks/**` | svc-tasks only |
| `apps/svc-weather/**` | svc-weather only |
| `apps/svc-reminders/**` | svc-reminders only |
| `apps/admin/**` | admin only |
| `apps/kiosk/**` | kiosk only |

### Key Behaviors
- **Shared-go changes**: Automatically triggers ALL Go service builds (they all depend on it)
- **Independent services**: Next.js apps (admin, kiosk) build independently
- **Deduplication**: If multiple triggers match the same service, it's only built once
- **No changes**: If no relevant files changed (e.g., README-only), no builds run

---

## Image Naming Convention

Published images follow this pattern:

```
ghcr.io/<owner>/home-hub-<service>:<tag>
```

### Examples
```bash
# Latest build
ghcr.io/jtumidanski/home-hub-svc-users:latest

# Specific commit
ghcr.io/jtumidanski/home-hub-svc-users:sha-a1b2c3d

# Semantic version (if Git tag exists)
ghcr.io/jtumidanski/home-hub-svc-users:v1.2.3
ghcr.io/jtumidanski/home-hub-admin:1.2
```

### Pulling Images
```bash
# Pull latest version
docker pull ghcr.io/<owner>/home-hub-svc-users:latest

# Pull specific commit
docker pull ghcr.io/<owner>/home-hub-svc-users:sha-a1b2c3d

# Pull specific version
docker pull ghcr.io/<owner>/home-hub-admin:v1.0.0
```

---

## Setup Requirements

### GitHub Secrets
No additional secrets required! The workflows use `GITHUB_TOKEN` which is automatically provided by GitHub Actions.

### Permissions
The `main-publish.yml` workflow requires:
- `contents: read` - To checkout code
- `packages: write` - To push images to GHCR

These are configured in the workflow file.

### Repository Settings
1. **Enable GitHub Actions**:
   - Settings → Actions → General → Allow all actions and reusable workflows

2. **Enable GHCR**:
   - Already enabled by default for all repositories

3. **Optional - Branch Protection**:
   - Settings → Branches → Add rule for `main`
   - Require status checks: Select "Pull Request Build"

---

## Build Caching

Both workflows use GitHub Actions cache for better performance:

- **Docker layer cache**: Speeds up subsequent builds
- **Scope**: Each service has its own cache scope
- **Mode**: `max` mode caches all layers

### Expected Performance
- **First build** (cold cache): 5-8 minutes per service
- **Subsequent builds** (warm cache): 2-4 minutes per service
- **Cache hit rate**: Typically 70-90%

---

## Testing Scenarios

### Test 1: Shared Library Change
```bash
# Make a change to shared-go
echo "// test" >> packages/shared-go/logger/logger.go
git checkout -b test-shared-go
git add .
git commit -m "Test shared-go change"
git push origin test-shared-go

# Create PR → Expect: All 4 Go services build
```

### Test 2: Single Service Change
```bash
# Make a change to one service
echo "// test" >> apps/svc-users/main.go
git checkout -b test-single-service
git add .
git commit -m "Test svc-users change"
git push origin test-single-service

# Create PR → Expect: Only svc-users builds
```

### Test 3: Frontend Change
```bash
# Make a change to admin
echo "// test" >> apps/admin/src/app/page.tsx
git checkout -b test-frontend
git add .
git commit -m "Test admin change"
git push origin test-frontend

# Create PR → Expect: Only admin builds
```

### Test 4: No Service Changes
```bash
# Change documentation only
echo "# test" >> README.md
git checkout -b test-docs
git add .
git commit -m "Update docs"
git push origin test-docs

# Create PR → Expect: No builds, workflow shows "No services changed"
```

---

## Troubleshooting

### Build Fails on PR

**Check**:
1. Does the Dockerfile exist at `apps/<service>/Dockerfile`?
2. Does the service compile locally with `docker build -f apps/<service>/Dockerfile .`?
3. Are all dependencies available (shared-go copied correctly)?

### Image Not Appearing in GHCR

**Check**:
1. Did the `main-publish.yml` workflow complete successfully?
2. Go to GitHub → Your Profile → Packages
3. Check package visibility settings (may be private by default)

### Service Not Building When Expected

**Check**:
1. Review the "Detect Changed Services" job output
2. Verify file paths match the patterns in `.github/actions/detect-changes/action.yml`
3. Check if the PR is against `main` or `master` branch

### Change Detection Not Working

**Debug**:
1. Go to Actions → Select failed run → "Detect Changed Services" job
2. Check the output logs for which files were detected as changed
3. Verify the file paths match the patterns defined in the action

### Manual Override

If auto-detection fails, you can:
1. **PR Build**: Use workflow_dispatch with "Force build all" option
2. **Main Publish**: Use workflow_dispatch with specific service names

---

## Maintenance

### Adding a New Service

When adding a new service:

1. **Update change detection** (`.github/actions/detect-changes/action.yml`):
   ```yaml
   # Add to files_yaml
   new_service:
     - apps/new-service/**

   # Add to detection logic
   if [[ "${{ steps.changed-files.outputs.new_service_any_changed }}" == "true" ]]; then
     echo "🔧 new-service changed"
     services+=("new-service")
   fi
   ```

2. **Test**: Create a test PR changing the new service

### Modifying Build Process

To customize the build:
- **Build arguments**: Add to `build-args` in both workflow files
- **Build platforms**: Add `platforms: linux/amd64,linux/arm64` for multi-platform
- **Cache configuration**: Modify `cache-from` and `cache-to` options

---

## Advanced Usage

### Multi-Platform Builds

To build for Raspberry Pi (arm64):

```yaml
- name: Set up QEMU
  uses: docker/setup-qemu-action@v3

- name: Build and push Docker image
  uses: docker/build-push-action@v5
  with:
    platforms: linux/amd64,linux/arm64
    # ... other options
```

### Semantic Versioning

To trigger semantic version tags:

```bash
# Create and push a Git tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3

# This will tag images with v1.2.3 and 1.2
```

---

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Build Push Action](https://github.com/docker/build-push-action)
- [GHCR Documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
- [Project Architecture](/docs/PROJECT_KNOWLEDGE.md)

---

**Last Updated**: 2025-11-14
