# Watered Plant Tracker - Justfile
# Run `just` to see all available commands

# Default recipe to display help
default:
    @just --list

# Development Commands
# ==================

# Run the development server
run:
    @echo "🚀 Starting Watered server..."
    @echo "💡 Demo login available at: http://localhost:8080/auth/demo-login"
    go run cmd/server/main.go

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

# Docker Commands (for future use)
# ===============================

# Build Docker image
docker-build:
    @echo "🐳 Building Docker image..."
    docker build -t watered:latest .

# Run with Docker
docker-run:
    @echo "🐳 Running with Docker..."
    docker run -p 8080:8080 --env-file .env watered:latest

# Docker compose up
docker-up:
    @echo "🐳 Starting with Docker Compose..."
    docker-compose up --build

# Docker compose down
docker-down:
    @echo "🐳 Stopping Docker Compose..."
    docker-compose down

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
    @echo "   1. Update .env with your Google OAuth credentials (see docs/GOOGLE_OAUTH_SETUP.md)"
    @echo "   2. Run 'just run' to start the server"
    @echo "   3. Visit http://localhost:8080/auth/demo-login for testing"

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