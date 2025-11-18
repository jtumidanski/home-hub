# GitHub Actions Quick Reference

## 🚀 Quick Commands

### Pull Latest Images
```bash
# Pull all services
docker pull ghcr.io/<owner>/home-hub-svc-users:latest
docker pull ghcr.io/<owner>/home-hub-svc-tasks:latest
docker pull ghcr.io/<owner>/home-hub-svc-weather:latest
docker pull ghcr.io/<owner>/home-hub-svc-reminders:latest
docker pull ghcr.io/<owner>/home-hub-admin:latest
docker pull ghcr.io/<owner>/home-hub-kiosk:latest
```

### Manual Workflow Triggers

**PR Build (Force All)**:
1. Go to Actions tab
2. Select "Pull Request Build"
3. Click "Run workflow"
4. Check "Force build all services"
5. Click "Run workflow"

**Main Publish (Specific Services)**:
1. Go to Actions tab
2. Select "Main Branch Publish"
3. Click "Run workflow"
4. Enter services: `svc-users,admin` (or leave empty for auto-detect)
5. Click "Run workflow"

---

## 📋 What Builds When?

| You Changed | What Builds |
|-------------|-------------|
| `packages/shared-go/**` | ✅ All 4 Go services |
| `go.work` | ✅ All 4 Go services |
| `apps/svc-users/**` | ✅ svc-users only |
| `apps/svc-tasks/**` | ✅ svc-tasks only |
| `apps/svc-weather/**` | ✅ svc-weather only |
| `apps/svc-reminders/**` | ✅ svc-reminders only |
| `apps/admin/**` | ✅ admin only |
| `apps/kiosk/**` | ✅ kiosk only |
| `README.md` | ⏭️ Nothing (skipped) |

---

## 🔧 Common Tasks

### Check Build Status
```bash
# View workflow runs
gh run list --workflow=pr-build.yml

# View specific run
gh run view <run-id>

# Watch live run
gh run watch
```

### View Published Images
```bash
# List all packages
gh api /user/packages?package_type=container

# View package versions
gh api /user/packages/container/home-hub-svc-users/versions
```

### Test Locally Before PR
```bash
# Build specific service
docker build -f apps/svc-users/Dockerfile -t test-svc-users .

# Run it
docker run --rm test-svc-users
```

---

## 🐛 Troubleshooting

### "No services changed" but I changed a file
**Fix**: Check if file path matches patterns in `.github/actions/detect-changes/action.yml`

### Build fails on PR but works locally
**Fix**:
1. Ensure you're building from repo root (context: `.`)
2. Check both service dir AND shared-go are copied in Dockerfile
3. Verify Go workspace replacement: `go mod edit -replace ...`

### Image not in GHCR
**Fix**:
1. Check workflow completed successfully
2. Go to github.com → Profile → Packages
3. Check package visibility (may be private)

### Wrong service built
**Fix**: Review "Detect Changed Services" job logs in workflow run

---

## ⚙️ Configuration

### No Extra Secrets Needed!
The workflows use `GITHUB_TOKEN` automatically provided by GitHub Actions.

### Optional: Branch Protection
Settings → Branches → main → Add rule:
- ✅ Require status checks to pass
- ✅ Select: "Pull Request Build"

---

## 📊 Performance

| Scenario | Expected Time |
|----------|---------------|
| Single service (cold) | ~5-8 min |
| Single service (warm) | ~2-4 min |
| All Go services (shared-go change) | ~15 min (parallel) |
| All services | ~20 min (parallel) |

Cache hit rate: 70-90% typical

---

For full documentation, see [.github/workflows/README.md](.github/workflows/README.md)
