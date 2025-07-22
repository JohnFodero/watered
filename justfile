# Watered Plant Tracker - Justfile
# Run `just` to see all available commands

# Default recipe to display help
default:
    @just --list

# Development Commands
# ==================

# Run the development server in demo mode (no Google auth, no env files loaded)
run:
    @echo "🐛 Starting Watered server in DEMO mode..."
    @echo "💡 Demo login available at: http://localhost:8080/auth/demo-login"
    @echo "🔐 Google OAuth DISABLED - using demo authentication"
    @echo "📊 Debug logs ENABLED"
    @echo "⚠️  No environment files loaded - demo users only"
    @echo ""
    WATERED_MODE=demo go run cmd/server/main.go

# Run the server in production mode (requires .env file with Google OAuth)
run-prod:
    @echo "🚀 Starting Watered server in PRODUCTION mode..."
    @echo "🔐 Google OAuth ENABLED - requires real authentication"
    @echo "🌍 Loading settings from .env files"
    @if [ ! -f .env ]; then echo "❌ .env file required for production mode. Copy .env.example to .env and configure it."; exit 1; fi
    @echo ""
    WATERED_MODE=production go run cmd/server/main.go

# Run the server with custom port
run-port PORT:
    @echo "🚀 Starting Watered server on port {{PORT}}..."
    PORT={{PORT}} go run cmd/server/main.go

# Stop any running Go servers
stop:
    @echo "🛑 Stopping running servers..."
    @pkill -f "go run cmd/server/main.go" || echo "No running servers found"
    @pkill -f "watered" || true

# Run development server with auto-reload (requires entr)
dev:
    @echo "🔄 Starting development server with auto-reload..."
    @echo "💡 Install entr for auto-reload: brew install entr"
    find . -name "*.go" -o -name "*.html" -o -name "*.css" | entr -r just run

# Testing Commands
# ===============

# Run all tests
test:
    @echo "🧪 Running all tests..."
    go test ./...

# Run tests with verbose output
test-verbose:
    @echo "🧪 Running all tests (verbose)..."
    go test -v ./...

# Run tests with coverage
test-coverage:
    @echo "🧪 Running tests with coverage..."
    go test -cover ./...
    @echo ""
    @echo "📊 Detailed coverage report:"
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "📝 Coverage report saved to coverage.html"

# Run tests for a specific package
test-package PACKAGE:
    @echo "🧪 Running tests for {{PACKAGE}}..."
    go test -v ./{{PACKAGE}}

# Build Commands
# =============

# Build the application
build:
    @echo "🔨 Building Watered application..."
    go build -o bin/watered cmd/server/main.go
    @echo "✅ Binary built: bin/watered"

# Build for multiple platforms
build-all:
    @echo "🔨 Building for multiple platforms..."
    @mkdir -p bin
    GOOS=darwin GOARCH=amd64 go build -o bin/watered-darwin-amd64 cmd/server/main.go
    GOOS=darwin GOARCH=arm64 go build -o bin/watered-darwin-arm64 cmd/server/main.go
    GOOS=linux GOARCH=amd64 go build -o bin/watered-linux-amd64 cmd/server/main.go
    GOOS=windows GOARCH=amd64 go build -o bin/watered-windows-amd64.exe cmd/server/main.go
    @echo "✅ Built for multiple platforms in bin/"

# Clean build artifacts
clean:
    @echo "🧹 Cleaning build artifacts..."
    rm -rf bin/
    rm -f coverage.out coverage.html
    go clean
    @echo "✅ Cleaned build artifacts"

# Code Quality Commands
# ====================

# Format Go code
fmt:
    @echo "🎨 Formatting Go code..."
    go fmt ./...
    @echo "✅ Code formatted"

# Tidy up go modules
tidy:
    @echo "📦 Tidying Go modules..."
    go mod tidy
    @echo "✅ Modules tidied"

# Run go vet for static analysis
vet:
    @echo "🔍 Running go vet..."
    go vet ./...
    @echo "✅ Static analysis complete"

# Check for common issues (fmt, vet, test)
check: fmt vet test
    @echo "✅ All checks passed!"

# Database Commands (for future use)
# =================================

# Initialize database (placeholder for future)
db-init:
    @echo "🗄️  Database initialization will be implemented in Task 4"

# Reset database (placeholder for future)
db-reset:
    @echo "🗄️  Database reset will be implemented in Task 4"

# Docker Commands
# ===============

# Build Docker image
docker-build:
    @echo "🐳 Building Docker image..."
    docker build -t watered:latest .
    @echo "✅ Docker image built: watered:latest"

# Build Docker image with no cache
docker-build-fresh:
    @echo "🐳 Building Docker image (no cache)..."
    docker build --no-cache -t watered:latest .
    @echo "✅ Docker image built: watered:latest"

# Google Cloud Artifact Registry Commands
# =======================================

# Setup Docker buildx for cross-platform builds
docker-setup-buildx:
    @echo "🔧 Setting up Docker buildx for cross-platform builds..."
    @if docker buildx inspect multiplatform >/dev/null 2>&1; then \
        echo "✅ Buildx builder 'multiplatform' already exists"; \
    else \
        echo "🏗️  Creating buildx builder for cross-platform builds..."; \
        docker buildx create --name multiplatform --use --bootstrap; \
        echo "✅ Buildx builder 'multiplatform' created and activated"; \
    fi
    @echo "🔍 Current buildx builder:"
    @docker buildx ls

# Build and tag image for Google Cloud Artifact Registry (AMD64 for compatibility)
docker-build-gcp:
    @echo "🐳 Building Docker image for Google Cloud (linux/amd64)..."
    @if [ -z "${GCP_PROJECT_ID}" ]; then echo "❌ GCP_PROJECT_ID environment variable not set"; exit 1; fi
    @if [ -z "${GCP_REGION}" ]; then echo "❌ GCP_REGION environment variable not set"; exit 1; fi
    @echo "🏗️  Building for linux/amd64 platform to ensure GCP compatibility..."
    docker buildx build --platform linux/amd64 -t watered:latest .
    docker tag watered:latest ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:latest
    docker tag watered:latest ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:$(git rev-parse --short HEAD)
    @echo "✅ Docker image built and tagged for GCP Artifact Registry (AMD64)"

# Push image to Google Cloud Artifact Registry
docker-push-gcp:
    @echo "🚀 Pushing Docker image to Google Cloud Artifact Registry..."
    @if [ -z "${GCP_PROJECT_ID}" ]; then echo "❌ GCP_PROJECT_ID environment variable not set"; exit 1; fi
    @if [ -z "${GCP_REGION}" ]; then echo "❌ GCP_REGION environment variable not set"; exit 1; fi
    @echo "🔐 Configuring Docker authentication for GCP..."
    gcloud auth configure-docker ${GCP_REGION}-docker.pkg.dev --quiet
    @echo "⬆️  Pushing latest tag..."
    docker push ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:latest
    @echo "⬆️  Pushing commit-specific tag..."
    docker push ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:$(git rev-parse --short HEAD)
    @echo "✅ Docker image pushed to GCP Artifact Registry"

# Build and push to Google Cloud Artifact Registry in one command
docker-deploy-gcp: docker-build-gcp docker-push-gcp
    @echo "🎉 Successfully deployed to Google Cloud Artifact Registry!"
    @echo "📦 Images available at:"
    @echo "   ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:latest"
    @echo "   ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:$(git rev-parse --short HEAD)"

# Setup Google Cloud authentication (interactive)
gcp-setup:
    @echo "🔧 Setting up Google Cloud configuration..."
    @echo "This will guide you through setting up authentication for GCP Artifact Registry"
    @echo ""
    @echo "1. Setting up Docker buildx for cross-platform builds..."
    @docker buildx inspect multiplatform >/dev/null 2>&1 || docker buildx create --name multiplatform --use --bootstrap
    @echo ""
    @echo "2. Authenticating with Google Cloud..."
    gcloud auth login
    @echo ""
    @echo "3. Configuring Docker authentication..."
    @if [ -z "${GCP_REGION}" ]; then \
        echo "Enter your GCP region (e.g., us-central1, europe-west1):"; \
        read -p "Region: " GCP_REGION; \
        echo "export GCP_REGION=$$GCP_REGION" >> .env.local; \
    fi
    @if [ -z "${GCP_PROJECT_ID}" ]; then \
        echo "Enter your GCP Project ID:"; \
        read -p "Project ID: " GCP_PROJECT_ID; \
        echo "export GCP_PROJECT_ID=$$GCP_PROJECT_ID" >> .env.local; \
    fi
    @echo ""
    @echo "✅ Google Cloud setup complete!"
    @echo "💡 Environment variables saved to .env.local"
    @echo "💡 Source them with: source .env.local"
    @echo "💡 Docker buildx configured for cross-platform builds"

# List images in Google Cloud Artifact Registry
gcp-list-images:
    @echo "📦 Listing images in Google Cloud Artifact Registry..."
    @if [ -z "${GCP_PROJECT_ID}" ]; then echo "❌ GCP_PROJECT_ID environment variable not set"; exit 1; fi
    @if [ -z "${GCP_REGION}" ]; then echo "❌ GCP_REGION environment variable not set"; exit 1; fi
    gcloud artifacts docker images list ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo

# Pull image from Google Cloud Artifact Registry
docker-pull-gcp TAG="latest":
    @echo "⬇️  Pulling Docker image from Google Cloud Artifact Registry..."
    @if [ -z "${GCP_PROJECT_ID}" ]; then echo "❌ GCP_PROJECT_ID environment variable not set"; exit 1; fi
    @if [ -z "${GCP_REGION}" ]; then echo "❌ GCP_REGION environment variable not set"; exit 1; fi
    gcloud auth configure-docker ${GCP_REGION}-docker.pkg.dev --quiet
    docker pull ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:{{TAG}}
    @echo "✅ Image pulled successfully"

# Run container from Google Cloud Artifact Registry
docker-run-gcp TAG="latest":
    @echo "🐳 Running container from Google Cloud Artifact Registry..."
    @if [ -z "${GCP_PROJECT_ID}" ]; then echo "❌ GCP_PROJECT_ID environment variable not set"; exit 1; fi
    @if [ -z "${GCP_REGION}" ]; then echo "❌ GCP_REGION environment variable not set"; exit 1; fi
    docker run -p 8080:8080 ${GCP_REGION}-docker.pkg.dev/${GCP_PROJECT_ID}/watered-repo/watered:{{TAG}}

# Run with Docker (development mode)
docker-run:
    @echo "🐳 Running with Docker..."
    @echo "💡 Demo login available at: http://localhost:8080/auth/demo-login?simple=true"
    docker run -p 8080:8080 watered:latest

# Run with Docker (production mode with env file)
docker-run-prod:
    @echo "🐳 Running with Docker (production mode)..."
    @if [ ! -f .env ]; then echo "❌ .env file not found! Create one first."; exit 1; fi
    docker run -p 8080:8080 --env-file .env watered:latest

# Docker compose up (development)
docker-up:
    @echo "🐳 Starting with Docker Compose (development mode)..."
    @echo "💡 Demo login available at: http://localhost:8080/auth/demo-login?simple=true"
    docker compose up --build

# Docker compose up (production with nginx)
docker-up-prod:
    @echo "🐳 Starting with Docker Compose (production mode)..."
    @echo "💡 App available at: http://localhost"
    docker compose --profile production up --build

# Docker compose up in background
docker-up-detached:
    @echo "🐳 Starting with Docker Compose (background)..."
    docker compose up --build -d
    @echo "✅ Services started in background"
    @echo "📜 View logs with: just docker-logs"
    @echo "🛑 Stop with: just docker-down"

# Docker compose down
docker-down:
    @echo "🐳 Stopping Docker Compose..."
    docker compose --profile production down
    @echo "✅ Services stopped"

# View Docker compose logs
docker-logs:
    @echo "📜 Viewing Docker Compose logs..."
    docker compose logs -f

# Restart Docker services
docker-restart:
    @echo "🔄 Restarting Docker services..."
    docker compose restart
    @echo "✅ Services restarted"

# Clean Docker resources
docker-clean:
    @echo "🧹 Cleaning Docker resources..."
    docker compose down --volumes --remove-orphans
    docker system prune -f
    @echo "✅ Docker resources cleaned"

# Show Docker status
docker-status:
    @echo "🐳 Docker Status"
    @echo "==============="
    @echo ""
    @echo "🖼️  Images:"
    @docker images | grep -E "(watered|nginx)" || echo "   No watered images found"
    @echo ""
    @echo "📦 Containers:"
    @docker ps -a | grep -E "(watered|nginx)" || echo "   No watered containers found"
    @echo ""
    @echo "🌐 Networks:"
    @docker network ls | grep watered || echo "   No watered networks found"

# Enter running Docker container
docker-shell:
    @echo "🐚 Entering running watered container..."
    @CONTAINER_ID=$$(docker ps -q -f "ancestor=watered:latest" | head -1); \
    if [ -z "$$CONTAINER_ID" ]; then \
        echo "❌ No running watered container found. Start with 'just docker-up'"; \
        exit 1; \
    else \
        docker exec -it $$CONTAINER_ID /bin/sh; \
    fi

# Test Docker container health
docker-health:
    @echo "🏥 Checking Docker container health..."
    @CONTAINER_ID=$$(docker ps -q -f "ancestor=watered:latest" | head -1); \
    if [ -z "$$CONTAINER_ID" ]; then \
        echo "❌ No running watered container found"; \
        exit 1; \
    else \
        echo "📊 Container health:"; \
        docker inspect $$CONTAINER_ID --format='{{{{.State.Health.Status}}}}' 2>/dev/null || echo "No health check configured"; \
        echo "🌐 Testing HTTP endpoint:"; \
        curl -s -o /dev/null -w "HTTP %{http_code} - %{time_total}s\n" http://localhost:8080/health || echo "❌ Health check failed"; \
    fi

# Setup Commands
# =============

# Initial project setup
setup:
    @echo "🏗️  Setting up Watered project..."
    @echo "1. Installing Go dependencies..."
    go mod download
    @echo "2. Creating directories..."
    @mkdir -p bin data logs
    @echo "3. Copying environment template..."
    @if [ ! -f .env ]; then cp .env.example .env; echo "📝 Created .env from template - update with your values!"; fi
    @echo "✅ Setup complete!"
    @echo ""
    @echo "📖 Next steps:"
    @echo "   1. Edit .env file with your configuration (see docs/env-configuration.md)"
    @echo "   2. For production: Add Google OAuth credentials to disable demo mode"
    @echo "   3. Run 'just run' to start the server (automatically loads .env)"
    @echo "   4. For demo mode: Visit http://localhost:8080/auth/demo-login"

# Install development dependencies
install-dev:
    @echo "🔧 Installing development dependencies..."
    @if command -v brew >/dev/null 2>&1; then \
        echo "Installing entr for auto-reload..."; \
        brew install entr || echo "entr already installed"; \
    else \
        echo "Homebrew not found - install entr manually for auto-reload"; \
    fi

# Utility Commands
# ===============

# Show project status
status:
    @echo "📊 Watered Project Status"
    @echo "========================="
    @go version
    @echo "Project root: `pwd`"
    @echo "Git branch: `git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'Not a git repo'`"
    @echo "Git commit: `git rev-parse --short HEAD 2>/dev/null || echo 'Not a git repo'`"
    @echo ""
    @echo "📁 File counts:"
    @echo "   Go files: `find . -name '*.go' | wc -l | tr -d ' '`"
    @echo "   Test files: `find . -name '*_test.go' | wc -l | tr -d ' '`"
    @echo "   HTML templates: `find . -name '*.html' | wc -l | tr -d ' '`"
    @echo ""
    @echo "🏃 Running processes:"
    @ps aux | grep "go run cmd/server/main.go" | grep -v grep || echo "   No server running"

# View server logs (if running in background)
logs:
    @echo "📜 Checking for running server..."
    @if pgrep -f "go run cmd/server/main.go" > /dev/null; then \
        echo "Server is running. Use Ctrl+C to stop."; \
    else \
        echo "No server currently running. Start with 'just run'"; \
    fi

# Open project in browser
open:
    @echo "🌐 Opening Watered in browser..."
    @if command -v open >/dev/null 2>&1; then \
        open http://localhost:8080; \
    elif command -v xdg-open >/dev/null 2>&1; then \
        xdg-open http://localhost:8080; \
    else \
        echo "Visit: http://localhost:8080"; \
    fi

# Git Commands
# ============

# Commit with a descriptive message
commit MESSAGE:
    @echo "📝 Committing changes..."
    git add .
    git commit -m "{{MESSAGE}}"

# Push to remote
push:
    @echo "⬆️  Pushing to remote..."
    git push

# Quick commit and push
save MESSAGE: (commit MESSAGE) push
    @echo "💾 Changes saved and pushed!"

# Security Commands
# ================

# Generate a secure session secret
generate-session-secret:
    @echo "🔐 Generating secure session secret..."
    @echo "Add this to your .env file:"
    @echo "SESSION_SECRET=`openssl rand -base64 32`"

# Check for sensitive files
security-check:
    @echo "🔒 Checking for sensitive files..."
    @if [ -f .env ]; then echo "⚠️  .env file exists - ensure it's not committed"; fi
    @git status --porcelain | grep -E "\\.env$$" && echo "⚠️  .env file is staged for commit!" || echo "✅ No .env file staged"
    @echo "✅ Security check complete"