# Google Cloud Platform Setup Guide

This guide will help you set up Google Cloud Artifact Registry integration for the Watered application, including local development and GitHub Actions CI/CD.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Google Cloud Project Setup](#google-cloud-project-setup)
- [Artifact Registry Setup](#artifact-registry-setup)
- [Local Development Setup](#local-development-setup)
- [GitHub Actions Setup](#github-actions-setup)
- [Usage](#usage)
- [Troubleshooting](#troubleshooting)

## Prerequisites

Before starting, ensure you have:

- A Google Cloud Platform account
- `gcloud` CLI installed and configured on your machine
- Docker installed and running
- Admin access to your GitHub repository
- The `just` command runner installed

### Install Google Cloud CLI

If you haven't already installed the Google Cloud CLI:

```bash
# On macOS
brew install google-cloud-sdk

# On Ubuntu/Debian
curl https://sdk.cloud.google.com | bash
exec -l $SHELL

# On Windows
# Download from: https://cloud.google.com/sdk/docs/install
```

## Google Cloud Project Setup

### 1. Create or Select a Project

```bash
# List existing projects
gcloud projects list

# Create a new project (replace PROJECT_ID with your desired ID)
gcloud projects create YOUR_PROJECT_ID --name="Watered App"

# Set the project as default
gcloud config set project YOUR_PROJECT_ID
```

### 2. Enable Required APIs

```bash
# Enable Artifact Registry API
gcloud services enable artifactregistry.googleapis.com

# Enable Cloud Build API (if you plan to use Cloud Build)
gcloud services enable cloudbuild.googleapis.com

# Verify enabled services
gcloud services list --enabled
```

### 3. Set Up Billing

Ensure your project has billing enabled in the [Google Cloud Console](https://console.cloud.google.com/billing).

## Artifact Registry Setup

### 1. Create the Artifact Registry Repository

```bash
# Set your region (choose the one closest to you)
export GCP_REGION="us-central1"  # or us-east1, europe-west1, etc.

# Create the repository
gcloud artifacts repositories create watered-repo \
    --repository-format=docker \
    --location=$GCP_REGION \
    --description="Docker images for Watered plant tracking app"

# Verify repository creation
gcloud artifacts repositories list --location=$GCP_REGION
```

### 2. Configure Repository Permissions

```bash
# Get your project number
PROJECT_NUMBER=$(gcloud projects describe YOUR_PROJECT_ID --format="value(projectNumber)")

# Grant the Compute Engine service account access (for Cloud Run, etc.)
gcloud artifacts repositories add-iam-policy-binding watered-repo \
    --location=$GCP_REGION \
    --member="serviceAccount:$PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
    --role="roles/artifactregistry.reader"
```

## Local Development Setup

### 1. Authenticate with Google Cloud

```bash
# Authenticate your gcloud CLI
gcloud auth login

# Set application default credentials
gcloud auth application-default login

# Configure Docker to use gcloud as a credential helper
gcloud auth configure-docker $GCP_REGION-docker.pkg.dev
```

### 2. Set Environment Variables

Create a `.env.local` file in your project root:

```bash
# .env.local
export GCP_PROJECT_ID="your-project-id"
export GCP_REGION="us-central1"  # Your chosen region
```

Source the file:

```bash
source .env.local
```

### 3. Setup Docker Buildx (Apple Silicon/ARM Macs)

**Important for Apple Silicon Macs:** Google Cloud services expect AMD64 (x86_64) images. If you're on an Apple Silicon Mac (M1/M2/M3), you need to set up Docker buildx for cross-platform builds:

```bash
# Setup buildx for cross-platform builds
just docker-setup-buildx
```

This command:
- Creates a Docker buildx builder named "multiplatform"
- Enables building for different architectures (AMD64, ARM64)
- Ensures compatibility with Google Cloud services

### 4. Test Local Setup

```bash
# Use the interactive setup command (includes buildx setup)
just gcp-setup

# Or manually test the setup
just docker-build-gcp  # Builds for linux/amd64 automatically
just docker-push-gcp
```

## GitHub Actions Setup

To enable automatic deployment to Google Cloud Artifact Registry from GitHub Actions, you need to set up authentication.

### 1. Create a Service Account

```bash
# Create a service account for GitHub Actions
gcloud iam service-accounts create github-actions \
    --description="Service account for GitHub Actions CI/CD" \
    --display-name="GitHub Actions"

# Get the service account email
SA_EMAIL="github-actions@YOUR_PROJECT_ID.iam.gserviceaccount.com"

# Grant necessary roles to the service account
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/artifactregistry.writer"

gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/storage.admin"
```

### 2. Create and Download Service Account Key

```bash
# Create a key file
gcloud iam service-accounts keys create github-actions-key.json \
    --iam-account=$SA_EMAIL

# Display the key content (you'll need this for GitHub secrets)
cat github-actions-key.json
```

**⚠️ Important:** Keep this key file secure and delete it from your local machine after setting up GitHub secrets.

### 3. Configure GitHub Repository Secrets

Go to your GitHub repository and navigate to **Settings → Secrets and variables → Actions**. Add the following repository secrets:

#### Required Secrets:

1. **`GCP_SA_KEY`**
   - Value: The entire content of the `github-actions-key.json` file
   - This authenticates GitHub Actions with Google Cloud

2. **`GCP_PROJECT_ID`**
   - Value: Your Google Cloud project ID
   - Example: `watered-app-123456`

#### Optional Secrets (if different from defaults):

3. **`GCP_REGION`** (if not using `us-central1`)
   - Value: Your chosen Google Cloud region
   - Example: `europe-west1`

### 4. Setting Up GitHub Secrets

#### Via GitHub Web Interface:

1. Go to `https://github.com/YOUR_USERNAME/YOUR_REPO/settings/secrets/actions`
2. Click "New repository secret"
3. Add each secret with the name and value as specified above

#### Via GitHub CLI:

```bash
# Set GCP_SA_KEY (the JSON key content)
gh secret set GCP_SA_KEY < github-actions-key.json

# Set GCP_PROJECT_ID
gh secret set GCP_PROJECT_ID --body "your-project-id"

# Set GCP_REGION (if needed)
gh secret set GCP_REGION --body "us-central1"
```

### 6. Additional Production Secrets (Optional)

For automated deployment of production-ready images with authentication disabled, you can also set:

```bash
# Production OAuth credentials (optional - for tagged production images)
gh secret set PROD_GOOGLE_CLIENT_ID --body "your-production-client-id"
gh secret set PROD_GOOGLE_CLIENT_SECRET --body "your-production-client-secret"
gh secret set PROD_SESSION_SECRET --body "your-production-session-secret"
gh secret set PROD_ALLOWED_EMAILS --body "you@domain.com,partner@domain.com"
gh secret set PROD_ADMIN_EMAILS --body "you@domain.com"
```

**Note**: The images pushed to Artifact Registry will still be in "demo mode" unless you configure OAuth credentials in your deployment environment (Cloud Run, Kubernetes, etc.). The container images are environment-agnostic and get their configuration from environment variables at runtime.

### 5. Clean Up Local Key File

```bash
# Delete the key file from your local machine
rm github-actions-key.json
```

## Usage

### Local Development

```bash
# Build and push to Google Cloud Artifact Registry
just docker-deploy-gcp

# List images in the registry
just gcp-list-images

# Pull and run an image from the registry
just docker-pull-gcp latest
just docker-run-gcp latest
```

### Automatic CI/CD

Once configured, the GitHub Actions workflow will automatically:

1. **On any push to `main` branch:**
   - Build the Docker image
   - Push to both GitHub Container Registry and Google Cloud Artifact Registry
   - Tag with both `latest` and the git commit SHA

2. **Image locations:**
   - GitHub: `ghcr.io/your-username/watered:latest`
   - Google Cloud: `us-central1-docker.pkg.dev/your-project/watered-repo/watered:latest`

### Manual Deployment

You can also manually trigger deployments:

```bash
# Build and push locally
just docker-deploy-gcp

# Or push a specific commit
git checkout COMMIT_SHA
just docker-deploy-gcp
```

## Advanced Configuration

### Custom Registry Names

If you want to use a different registry name, update the following:

1. **Justfile:** Update the registry name in the GCP commands
2. **CI/CD:** Update the `images` field in the workflow file
3. **GCP:** Create the repository with your desired name

### Multi-Region Setup

To deploy to multiple regions:

```bash
# Create repositories in multiple regions
for region in us-central1 europe-west1 asia-east1; do
    gcloud artifacts repositories create watered-repo \
        --repository-format=docker \
        --location=$region \
        --description="Docker images for Watered plant tracking app"
done
```

### Repository Cleanup Policies

Set up automatic cleanup of old images:

```bash
# Create a cleanup policy (keep only 10 latest images)
gcloud artifacts repositories create watered-repo \
    --repository-format=docker \
    --location=$GCP_REGION \
    --cleanup-policy-file=cleanup-policy.json
```

Create `cleanup-policy.json`:

```json
{
  "name": "keep-recent-images",
  "condition": {
    "tagState": "TAGGED",
    "newerThan": "2592000s"
  },
  "action": {
    "type": "Keep"
  },
  "mostRecentVersions": {
    "keepCount": 10
  }
}
```

## Troubleshooting

### Common Issues

#### Platform Compatibility Issues (Apple Silicon Macs)

```bash
# Error: "exec /app/watered: exec format error" in Google Cloud
# or: Container fails to start with architecture mismatch
# Cause: ARM64 image built on Apple Silicon Mac incompatible with GCP AMD64 infrastructure

# Solution 1: Use the GCP-specific build command (builds for AMD64)
just docker-build-gcp

# Solution 2: Setup buildx if not already done
just docker-setup-buildx

# Solution 3: Manually specify platform
docker buildx build --platform linux/amd64 -t watered:latest .

# Verify image architecture
docker inspect watered:latest | grep Architecture
# Should show: "Architecture": "amd64"
```

#### Authentication Errors

```bash
# Error: "access denied" or "permission denied"
# Solution: Re-authenticate
gcloud auth login
gcloud auth configure-docker $GCP_REGION-docker.pkg.dev
```

#### Repository Not Found

```bash
# Error: "repository does not exist"
# Solution: Verify repository exists
gcloud artifacts repositories list --location=$GCP_REGION
```

#### Docker Buildx Issues

```bash
# Error: "buildx: unknown builder instance"
# Solution: Setup or recreate buildx builder
just docker-setup-buildx

# Error: "multiple platforms feature is currently not supported"
# Solution: Enable experimental features or use newer Docker version
export DOCKER_CLI_EXPERIMENTAL=enabled
docker version
```

#### GitHub Actions Fails

Common GitHub Actions issues:

1. **Invalid service account key:** Verify `GCP_SA_KEY` secret is correctly set
2. **Wrong project ID:** Verify `GCP_PROJECT_ID` secret matches your project
3. **API not enabled:** Ensure Artifact Registry API is enabled
4. **Platform mismatch:** GitHub Actions builds for AMD64 by default (correct for GCP)

#### Local Environment Issues

```bash
# Reset Docker authentication
gcloud auth configure-docker $GCP_REGION-docker.pkg.dev --quiet

# Check current configuration
gcloud config list

# Verify project and authentication
gcloud auth list
gcloud config get-value project

# Check buildx status
docker buildx ls
```

### Debugging Commands

```bash
# Test Docker authentication
docker pull $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/watered-repo/watered:latest

# Check repository permissions
gcloud artifacts repositories get-iam-policy watered-repo --location=$GCP_REGION

# View recent pushes
gcloud artifacts docker images list $GCP_REGION-docker.pkg.dev/$GCP_PROJECT_ID/watered-repo

# Check service account permissions
gcloud projects get-iam-policy $GCP_PROJECT_ID \
    --filter="bindings.members:serviceAccount:github-actions@$GCP_PROJECT_ID.iam.gserviceaccount.com"
```

### Getting Help

If you encounter issues:

1. Check the [Google Cloud Artifact Registry documentation](https://cloud.google.com/artifact-registry/docs)
2. Verify all APIs are enabled in your project
3. Ensure billing is enabled on your Google Cloud project
4. Check GitHub Actions logs for detailed error messages
5. Verify all environment variables and secrets are correctly set

### Cost Considerations

- **Storage costs:** ~$0.10 per GB per month
- **Bandwidth costs:** ~$0.12 per GB egress to internet
- **Operations costs:** Minimal for typical usage

Monitor costs in the [Google Cloud Console](https://console.cloud.google.com/billing) billing section.

## Security Best Practices

1. **Limit service account permissions:** Only grant necessary roles
2. **Rotate service account keys:** Regularly rotate keys used by GitHub Actions
3. **Use least privilege:** Don't grant broader permissions than needed
4. **Monitor access:** Review audit logs regularly
5. **Secure secrets:** Never commit service account keys to version control

## Next Steps

After setting up GCP integration:

1. **Set up Cloud Run:** Deploy containers directly from Artifact Registry
2. **Configure monitoring:** Set up alerting for image pushes and pulls
3. **Implement Blue/Green deployments:** Use multiple image tags for safe deployments
4. **Set up vulnerability scanning:** Enable automatic scanning of images
5. **Configure backup:** Set up cross-region replication for disaster recovery