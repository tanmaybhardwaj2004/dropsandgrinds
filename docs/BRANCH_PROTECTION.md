# Branch Protection Rules Setup

Branch protection rules must be configured manually in GitHub repository settings.

## Required Branch Protection Rules for `main` Branch

Navigate to: Repository Settings → Branches → Branch protection rules → Add rule

### Rule Name: `main`

**Required Settings:**

1. **Require status checks to pass before merging**
   - Enable: ✅
   - Required checks:
     - `Go Test And Vet`
     - `Trivy Docker Image Scan`

2. **Require branches to be up to date before merging**
   - Enable: ✅

3. **Require pull request reviews before merging**
   - Enable: ✅ (optional but recommended)
   - Required approving reviews: 1
   - Dismiss stale PR approvals when new commits are pushed: ✅

4. **Require conversation resolution before merging**
   - Enable: ✅ (optional but recommended)

5. **Restrict who can push to matching branches**
   - Enable: ✅
   - Allow: Only administrators and specific users

6. **Do not allow bypassing the above settings**
   - Enable: ✅

## Why These Rules?

- **Go Test And Vet**: Ensures all Go tests pass and code is vetted before merging
- **Trivy Docker Image Scan**: Blocks merges with HIGH/CRITICAL security vulnerabilities
- **Up to date**: Ensures PRs are based on the latest `main` to avoid conflicts
- **Review approval**: Ensures code review before production deployment
- **Conversation resolution**: Ensures all discussions are resolved before merging
- **Push restrictions**: Prevents force pushes and unauthorized direct commits
- **No bypass**: Ensures rules cannot be circumvented

## Setup Steps

1. Go to your GitHub repository
2. Click Settings → Branches
3. Click "Add branch protection rule"
4. Enter `main` as the branch name pattern
5. Configure the settings above
6. Click "Create"

## CI/CD Integration

The `.github/workflows/deploy.yml` workflow is configured to:
- Run tests and vet on all branches
- Run security scan on all branches
- Only build and push to ECR on `main` branch
- Trigger Watchtower deployment on `main` branch

With branch protection enabled, the workflow must pass before any PR can be merged to `main`.
