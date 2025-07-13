# Just Commands Reference

This document provides a comprehensive reference for all available `just` commands in the Watered project.

## Quick Start

```bash
# View all available commands
just

# Setup Google Cloud integration
just gcp-setup

# Build and deploy to Google Cloud
just docker-deploy-gcp
```

## Google Cloud Artifact Registry Commands

### Prerequisites

Set up your environment variables:

```bash
export GCP_PROJECT_ID="your-project-id"
export GCP_REGION="us-central1"  # or your preferred region
```

### Commands

| Command | Description | Example |
|---------|-------------|---------|
| `just docker-setup-buildx` | Setup Docker buildx for cross-platform builds | `just docker-setup-buildx` |
| `just gcp-setup` | Interactive setup for Google Cloud authentication | `just gcp-setup` |
| `just docker-build-gcp` | Build and tag Docker image for GCP (AMD64) | `just docker-build-gcp` |
| `just docker-push-gcp` | Push Docker image to GCP Artifact Registry | `just docker-push-gcp` |
| `just docker-deploy-gcp` | Build and push in one command | `just docker-deploy-gcp` |
| `just gcp-list-images` | List all images in the registry | `just gcp-list-images` |
| `just docker-pull-gcp [TAG]` | Pull image from GCP registry | `just docker-pull-gcp latest` |
| `just docker-run-gcp [TAG]` | Run container from GCP registry | `just docker-run-gcp latest` |

### Environment Variables

The GCP commands require these environment variables:

- `GCP_PROJECT_ID`: Your Google Cloud project ID
- `GCP_REGION`: Your Google Cloud region (e.g., `us-central1`)

### Examples

```bash
# Initial setup (especially important for Apple Silicon Macs)
export GCP_PROJECT_ID="watered-app-123456"
export GCP_REGION="us-central1"

# Setup buildx for cross-platform builds (Apple Silicon/ARM Macs)
just docker-setup-buildx

# Build and deploy (automatically builds for AMD64)
just docker-deploy-gcp

# List deployed images
just gcp-list-images

# Pull and run a specific version
just docker-pull-gcp main-abc1234
just docker-run-gcp main-abc1234
```

## All Available Commands

### Development Commands

| Command | Description |
|---------|-------------|
| `just run` | Start the development server |
| `just run-port PORT` | Start server on custom port |
| `just stop` | Stop running servers |
| `just dev` | Start with auto-reload (requires entr) |

### Testing Commands

| Command | Description |
|---------|-------------|
| `just test` | Run all tests |
| `just test-verbose` | Run tests with verbose output |
| `just test-coverage` | Run tests with coverage report |
| `just test-package PACKAGE` | Run tests for specific package |

### Build Commands

| Command | Description |
|---------|-------------|
| `just build` | Build the application binary |
| `just build-all` | Build for multiple platforms |
| `just clean` | Clean build artifacts |

### Code Quality Commands

| Command | Description |
|---------|-------------|
| `just fmt` | Format Go code |
| `just tidy` | Tidy Go modules |
| `just vet` | Run static analysis |
| `just check` | Run fmt, vet, and test |

### Docker Commands

| Command | Description |
|---------|-------------|
| `just docker-build` | Build Docker image |
| `just docker-build-fresh` | Build with no cache |
| `just docker-run` | Run with Docker |
| `just docker-run-prod` | Run in production mode |
| `just docker-up` | Start with Docker Compose |
| `just docker-up-prod` | Start production compose |
| `just docker-up-detached` | Start in background |
| `just docker-down` | Stop Docker services |
| `just docker-logs` | View Docker logs |
| `just docker-restart` | Restart Docker services |
| `just docker-clean` | Clean Docker resources |
| `just docker-status` | Show Docker status |
| `just docker-shell` | Enter running container |
| `just docker-health` | Check container health |

### Setup Commands

| Command | Description |
|---------|-------------|
| `just setup` | Initial project setup |
| `just install-dev` | Install development dependencies |

### Utility Commands

| Command | Description |
|---------|-------------|
| `just status` | Show project status |
| `just logs` | View server logs |
| `just open` | Open project in browser |

### Git Commands

| Command | Description |
|---------|-------------|
| `just commit MESSAGE` | Commit with message |
| `just push` | Push to remote |
| `just save MESSAGE` | Commit and push |

### Security Commands

| Command | Description |
|---------|-------------|
| `just generate-session-secret` | Generate secure session secret |
| `just security-check` | Check for sensitive files |

## Workflow Examples

### Local Development

```bash
# Initial setup
just setup
just install-dev

# Start development
just dev  # Auto-reload enabled
# OR
just run  # Simple start

# Code quality check
just check
```

### Docker Development

```bash
# Build and run with Docker
just docker-build
just docker-run

# Or use Docker Compose
just docker-up

# Check health
just docker-health

# View logs
just docker-logs

# Clean up
just docker-down
just docker-clean
```

### Production Deployment

```bash
# Build for production
just build-all

# Deploy to Google Cloud
export GCP_PROJECT_ID="your-project"
export GCP_REGION="us-central1"
just docker-deploy-gcp

# Or use production Docker Compose
just docker-up-prod
```

### Testing Workflow

```bash
# Run all tests
just test

# Generate coverage report
just test-coverage

# Test specific package
just test-package internal/auth
```

### CI/CD Integration

The GitHub Actions workflow automatically uses these commands:

- `just test` - Run tests
- `just build` - Build application
- `just docker-build-gcp` - Build for GCP (on main branch)
- `just docker-push-gcp` - Push to GCP (on main branch)

## Configuration

### Environment Files

- `.env` - Main environment configuration
- `.env.local` - Local GCP configuration (git-ignored)
- `.env.example` - Template with all variables

### Required for GCP Commands

```bash
# Add to .env or .env.local
GCP_PROJECT_ID=your-project-id
GCP_REGION=us-central1
```

### Google Cloud Authentication

```bash
# Authenticate locally
gcloud auth login
gcloud auth application-default login

# Configure Docker
gcloud auth configure-docker us-central1-docker.pkg.dev
```

## Tips and Best Practices

1. **Use `just` without arguments** to see all available commands
2. **Set up `.env.local`** for GCP variables to avoid committing them
3. **Use `just dev`** for development with auto-reload
4. **Run `just check`** before committing code
5. **Use `just docker-health`** to verify container status
6. **Use `just security-check`** to avoid committing sensitive files
7. **Use specific tags** when pulling/running GCP images in production
8. **Apple Silicon Macs**: Always run `just docker-setup-buildx` first for GCP compatibility
9. **Platform compatibility**: Use `just docker-build-gcp` instead of `just docker-build` for GCP deployments

## Troubleshooting

### Common Issues

1. **GCP commands fail**: Ensure `GCP_PROJECT_ID` and `GCP_REGION` are set
2. **Docker authentication fails**: Run `gcloud auth configure-docker`
3. **Server won't start**: Check if port 8080 is already in use
4. **Tests fail**: Ensure all dependencies are installed with `just setup`

### Getting Help

```bash
# View command help
just --help

# List all commands
just --list

# View project status
just status
```