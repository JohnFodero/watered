# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

"Watered" is a simple web application for tracking plant watering between partners. Users log in via Google SSO to see a visual plant representation that changes based on watering status. Tapping the plant resets the watering timer. Includes an admin panel for managing timeout settings and authorized email addresses.

**Tech Stack:**
- Frontend: Alpine.js, HTML/CSS/JavaScript
- Backend: Go (Golang)
- Authentication: Google SSO
- Deployment: Docker (self-hosted initially)
- Future: Nix shells, cloud hosting

## Development Guidelines

**Core Principles:**
- Keep it simple first, add complexity feature by feature
- Test everything - write tests for all functionality
- Document setup and testing procedures thoroughly
- Focus on learning frontend development and Go skills

**Coding Conventions:**
- Write comprehensive tests for all new features
- Follow Go best practices and idiomatic patterns
- Use Alpine.js for reactive frontend behavior
- Maintain clean, readable code structure
- Always make a new commit after a task is completed and all tests for that step are passing

**Development Workflow:**
- **ALWAYS use `just` commands** for building, testing, and running the application
- **Test thoroughly** before committing - run `just test` to ensure all functionality works
- **Write tests for new features** - add unit tests, integration tests, and e2e tests as needed
- **Validate with multiple test types** - ensure code works in isolation and in full system context
- **Fix failing tests immediately** - never commit with failing tests unless explicitly documenting why

## Development Setup

### Prerequisites
- Go 1.21+
- Docker and Docker Compose
- Node.js (for any frontend tooling)

### Running the Application
**Use `just` commands for all development tasks - run `just` to see all available commands.**

```bash
# Primary development commands
just run              # Run the development server
just test             # Run all tests
just build            # Build the application
just check            # Run formatting, linting, and tests

# Development workflow
just dev              # Run with auto-reload (requires entr)
just docker-up        # Run with Docker Compose
just test-verbose     # Run tests with detailed output
just test-coverage    # Run tests with coverage report

# Code quality
just fmt              # Format code
just vet              # Run static analysis
just tidy             # Clean up module dependencies
```

**Important:** Always use `just test` before committing. Add new tests for any functionality you create.

### Project Structure
```
watered/
├── cmd/server/          # Application entrypoint
├── internal/           # Private application code
│   ├── auth/          # Authentication logic
│   ├── handlers/      # HTTP handlers
│   ├── models/        # Data models
│   ├── services/      # Business logic
│   ├── storage/       # Database layer
│   └── monitoring/    # Health checks and monitoring
├── web/               # Frontend assets
│   ├── static/        # CSS, JS, images
│   └── templates/     # HTML templates
├── tests/             # Test files
│   ├── e2e/          # End-to-end tests
│   ├── integration/  # Integration tests
│   └── performance/  # Performance tests
├── justfile           # Build and development commands
└── docker-compose.yml # Development environment
```

## Architecture

**Frontend:**
- Single-page application using Alpine.js for reactivity
- Static HTML templates served by Go backend
- Plant visualization with CSS animations/transitions
- Mobile-responsive design

**Backend:**
- Go HTTP server with Chi router (github.com/go-chi/chi/v5)
- RESTful API for plant state management
- Google OAuth2 integration for authentication
- In-memory storage (currently), SQLite planned for persistence
- Admin API endpoints for configuration
- Comprehensive health monitoring and logging

**Authentication Flow:**
1. User visits app, redirected to Google OAuth
2. Google returns with user info and token
3. Backend validates token and checks email whitelist
4. Session established with secure cookies
5. Frontend receives authentication state

**Data Models:**
- User (email, name, admin status)
- PlantState (last_watered, timeout_hours)
- AdminConfig (timeout, whitelisted_emails)

## Testing Strategy

**Test Coverage Requirements:**
- **Unit Tests:** Test individual functions and methods in isolation
- **Integration Tests:** Test API endpoints and service interactions
- **End-to-End Tests:** Test complete user workflows
- **Performance Tests:** Validate system performance under load

**Testing Commands:**
```bash
just test              # Run all tests
just test-verbose      # Run with detailed output
just test-coverage     # Generate coverage report
just test-package auth # Test specific package
```

**Testing Guidelines:**
- Write tests for all new features before implementing them (TDD approach)
- Ensure tests are fast, reliable, and independent
- Mock external dependencies (Google OAuth, etc.)
- Test error conditions and edge cases
- Maintain high test coverage (aim for >80%)
- Use table-driven tests for multiple scenarios
- Add integration tests for API endpoints
- Include e2e tests for critical user flows

## Design Theme

**Color Palette (Monokai Pro Light Filter Sun):**
- Primary Background: `#f8efe7` (soft warm beige)
- Secondary Background: `#ede5de` (slightly darker warm neutral)
- Activity Background: `#ded5d0` (muted rose-taupe)
- Primary Text: `#2c232d` (deep plum-brown)
- Accent Color: `#cd4770` (vibrant rose-pink)
- Muted Text: `#91898a` (soft gray-mauve)

**Design Principles:**
- Soft, warm color scheme with delicate neutrals
- Gentle, nurturing interface that feels organic
- Subtle animations for plant state transitions
- Clean, minimalist design focused on usability